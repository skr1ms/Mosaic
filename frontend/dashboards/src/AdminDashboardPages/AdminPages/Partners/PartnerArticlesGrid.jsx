import React, { useState, useEffect, useCallback } from 'react'
import { X, Check, AlertCircle, ExternalLink } from 'lucide-react'
import { useParams, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { toast } from 'react-toastify'
import axios from 'axios'

const PartnerArticlesGrid = () => {
  const { partnerId } = useParams()
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [articles, setArticles] = useState({})
  const [editingCell, setEditingCell] = useState(null)
  const [tempValue, setTempValue] = useState('')
  const [partner, setPartner] = useState(null)

  const [savedCells, setSavedCells] = useState(new Set())
  const [focusedCell, setFocusedCell] = useState(null)

  const sizes = [
    { key: '21x30', title: '21×30' },
    { key: '30x40', title: '30×40' },
    { key: '40x40', title: '40×40' },
    { key: '40x50', title: '40×50' },
    { key: '40x60', title: '40×60' },
    { key: '50x70', title: '50×70' }
  ]

  const styles = [
    { key: 'grayscale', title: t('partners.styles.grayscale') },
    { key: 'skin_tones', title: t('partners.styles.skin_tones') },
    { key: 'pop_art', title: t('partners.styles.pop_art') },
    { key: 'max_colors', title: t('partners.styles.max_colors') }
  ]

  const marketplaces = [
    { key: 'ozon', title: 'OZON', color: 'bg-orange-100 text-orange-800' },
    { key: 'wildberries', title: 'Wildberries', color: 'bg-purple-100 text-purple-800' }
  ]

  const fetchPartnerData = useCallback(async () => {
    try {
      const token = localStorage.getItem('adminToken')
      const response = await axios.get(
        `${import.meta.env.VITE_API_URL}/api/admin/partners/${partnerId}`,
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      )
      setPartner(response.data)
    } catch (error) {
      console.error('Error fetching partner:', error)
      toast.error(t('partners.error_loading_partner'))
    }
  }, [partnerId, t])

  const fetchArticles = useCallback(async () => {
    try {
      const token = localStorage.getItem('adminToken')

      const response = await axios.get(
        `${import.meta.env.VITE_API_URL}/api/admin/partners/${partnerId}/articles/grid`,
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      )

      
      
      const articlesMap = {}
      const gridData = response.data || {}

      Object.keys(gridData).forEach(marketplace => {
        const marketplaceData = gridData[marketplace] || {}
        Object.keys(marketplaceData).forEach(style => {
          const styleData = marketplaceData[style] || {}
          Object.keys(styleData).forEach(size => {
            const key = `${marketplace}-${style}-${size}`
            articlesMap[key] = styleData[size] || ''
          })
        })
      })

      setArticles(articlesMap)

      setLoading(false)
    } catch (error) {
      console.error('Error fetching articles:', error)
      toast.error(t('partners.error_loading_articles'))
      setLoading(false)
    }
  }, [partnerId, t])

  useEffect(() => {
    fetchPartnerData()
    fetchArticles()
  }, [partnerId, fetchPartnerData, fetchArticles])

  const handleCellClick = (marketplace, style, size) => {
    const key = `${marketplace}-${style}-${size}`

    
    if (editingCell && editingCell !== key) {
      handleSaveCell()
    }

    
    setEditingCell(key)
    setTempValue(articles[key] || '')
  }

  // Save article directly to database
  const handleSaveCell = async () => {
    if (!editingCell || saving) return

    const [marketplace, style, size] = editingCell.split('-')
    const sku = tempValue.trim()

    try {
      setSaving(true)
      const token = localStorage.getItem('adminToken')

      await axios.put(
        `${import.meta.env.VITE_API_URL}/api/admin/partners/${partnerId}/articles/sku`,
        { size, style, marketplace, sku },
        { headers: { Authorization: `Bearer ${token}` } }
      )

      
      setArticles(prev => ({
        ...prev,
        [editingCell]: sku
      }))

      
      if (sku) {
        setSavedCells(prev => new Set(prev).add(editingCell))
      } else {
        setSavedCells(prev => {
          const newSet = new Set(prev)
          newSet.delete(editingCell)
          return newSet
        })
      }

      setEditingCell(null)
      setTempValue('')

    } catch (error) {
      console.error('Error saving article:', error)
      toast.error('Ошибка при сохранении артикула')
    } finally {
      setSaving(false)
    }
  }



  const handleGenerateURL = async (marketplace, style, size) => {
    try {
      const token = localStorage.getItem('adminToken')
      const response = await axios.post(
        `${import.meta.env.VITE_API_URL}/api/admin/partners/${partnerId}/articles/generate-url`,
        {
          marketplace,
          style,
          size
        },
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      )

      const { url, has_article } = response.data

      if (has_article && url) {
        
        await navigator.clipboard.writeText(url)

        
        window.open(url, '_blank')
      } else {
        toast.warning(t('partners.no_sku_found'))
      }
    } catch (error) {
      console.error('Error generating URL:', error)
      toast.error(t('partners.failed_to_generate_url'))
    }
  }

  const handleCancelEdit = () => {
    setEditingCell(null)
    setTempValue('')
  }

  // Navigation functions
  const getNextCell = (currentMarketplace, currentStyle, currentSize, direction) => {
    const marketplaceIndex = marketplaces.findIndex(m => m.key === currentMarketplace)
    const styleIndex = styles.findIndex(s => s.key === currentStyle)
    const sizeIndex = sizes.findIndex(s => s.key === currentSize)

    let newMarketplace = currentMarketplace
    let newStyle = currentStyle
    let newSize = currentSize

    switch (direction) {
      case 'ArrowUp':
        if (styleIndex > 0) {
          newStyle = styles[styleIndex - 1].key
        } else if (marketplaceIndex > 0) {
          newMarketplace = marketplaces[marketplaceIndex - 1].key
          newStyle = styles[styles.length - 1].key
        }
        break
      case 'ArrowDown':
        if (styleIndex < styles.length - 1) {
          newStyle = styles[styleIndex + 1].key
        } else if (marketplaceIndex < marketplaces.length - 1) {
          newMarketplace = marketplaces[marketplaceIndex + 1].key
          newStyle = styles[0].key
        }
        break
      case 'ArrowLeft':
        if (sizeIndex > 0) {
          newSize = sizes[sizeIndex - 1].key
        } else if (styleIndex > 0) {
          newStyle = styles[styleIndex - 1].key
          newSize = sizes[sizes.length - 1].key
        } else if (marketplaceIndex > 0) {
          newMarketplace = marketplaces[marketplaceIndex - 1].key
          newStyle = styles[styles.length - 1].key
          newSize = sizes[sizes.length - 1].key
        }
        break
      case 'ArrowRight':
        if (sizeIndex < sizes.length - 1) {
          newSize = sizes[sizeIndex + 1].key
        } else if (styleIndex < styles.length - 1) {
          newStyle = styles[styleIndex + 1].key
          newSize = sizes[0].key
        } else if (marketplaceIndex < marketplaces.length - 1) {
          newMarketplace = marketplaces[marketplaceIndex + 1].key
          newStyle = styles[0].key
          newSize = sizes[0].key
        }
        break
      default:
        break
    }

    return { marketplace: newMarketplace, style: newStyle, size: newSize }
  }

  const handleKeyNavigation = (e, currentMarketplace, currentStyle, currentSize) => {
    if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Tab'].includes(e.key)) {
      e.preventDefault()
      e.stopPropagation()

      let direction = e.key
      if (e.key === 'Tab') {
        direction = e.shiftKey ? 'ArrowLeft' : 'ArrowRight'
      }

      const nextCell = getNextCell(currentMarketplace, currentStyle, currentSize, direction)
      const hasChanged = nextCell.marketplace !== currentMarketplace || nextCell.style !== currentStyle || nextCell.size !== currentSize

      if (hasChanged) {
        
        const newFocusedCell = `${nextCell.marketplace}-${nextCell.style}-${nextCell.size}`
        setFocusedCell(newFocusedCell)

        
        setTimeout(() => {
          const nextCellButton = document.querySelector(`button[data-cell="${newFocusedCell}"]`)
          if (nextCellButton) {
            nextCellButton.focus()
          }
        }, 100)
      }
    }

    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      e.stopPropagation()
      handleCellClick(currentMarketplace, currentStyle, currentSize)
    }
  }

  const getCellValue = (marketplace, style, size) => {
    const key = `${marketplace}-${style}-${size}`
    return articles[key] || ''
  }

  const isEditing = (marketplace, style, size) => {
    const key = `${marketplace}-${style}-${size}`
    return editingCell === key
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">{t('partners.loading_articles')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      {}
      <div className="mb-6">
        <button
          onClick={() => navigate('/admin/partners')}
          className="text-blue-600 hover:text-blue-800 mb-4 inline-flex items-center"
        >
          {t('partners.back_to_partners')}
        </button>

        <div className="bg-white rounded-lg shadow p-6">
          <h1 className="text-4xl font-bold text-gray-900 mb-6 text-center">
            {t('partners.article_grid_title')}
          </h1>
          {partner && (
            <p className="text-gray-600 text-lg text-center mb-4">
              {t('partners.partner_label')}: <span className="font-semibold text-gray-900">{partner.brand_name}</span>
            </p>
          )}

          <div className="mt-4 p-4 bg-blue-50 rounded-lg">
            <div className="flex items-start">
              <AlertCircle className="w-5 h-5 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
              <div className="text-sm text-blue-800">
                <p className="font-semibold mb-1">{t('partners.article_grid_instructions_title')}</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>{t('partners.article_grid_instruction_1')}</li>
                  <li>{t('partners.article_grid_instruction_2')}</li>
                  <li>{t('partners.article_grid_instruction_3')}</li>
                  <li>{t('partners.article_grid_instruction_4')}</li>
                  <li>{t('partners.article_grid_instruction_5')}</li>
                  <li>{t('partners.article_grid_instruction_6')}</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>

      {}
      {marketplaces.map(marketplace => (
        <div key={marketplace.key} className="mb-8">
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className={`px-6 py-4 ${marketplace.color}`}>
              <h2 className="text-xl font-semibold">
                {t('partners.marketplace_article_table')} {marketplace.title}
              </h2>
            </div>

            <div className="p-6">
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead>
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        {t('partners.style')} / {t('partners.size')}
                      </th>
                      {sizes.map(size => (
                        <th key={size.key} className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                          {size.title}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {styles.map(style => (
                      <tr key={style.key}>
                        <td className="px-4 py-3 text-sm font-medium text-gray-900 whitespace-nowrap">
                          {style.title}
                        </td>
                        {sizes.map(size => {
                          const cellKey = `${marketplace.key}-${style.key}-${size.key}`
                          const isEditingCell = isEditing(marketplace.key, style.key, size.key)
                          const value = getCellValue(marketplace.key, style.key, size.key)
                          const isSaved = savedCells.has(cellKey) || (value && !isEditingCell)

                          return (
                            <td key={size.key} className="px-2 py-2">
                              {isEditingCell ? (
                                <div className="flex items-center space-x-1">
                                  <input
                                    type="text"
                                    value={tempValue}
                                    onChange={(e) => setTempValue(e.target.value)}
                                    className="w-full px-2 py-1 text-sm border border-blue-500 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    placeholder={t('partners.enter_sku')}
                                    autoFocus
                                    onBlur={handleSaveCell}
                                    onKeyDown={(e) => {
                                      if (e.key === 'Enter') {
                                        handleSaveCell()
                                        return
                                      }
                                      if (e.key === 'Escape') {
                                        handleCancelEdit()
                                        return
                                      }
                                      if (e.key === 'Tab') {
                                        e.preventDefault()
                                        e.stopPropagation()
                                        handleSaveCell()

                                        const direction = e.shiftKey ? 'ArrowLeft' : 'ArrowRight'
                                        const nextCell = getNextCell(marketplace.key, style.key, size.key, direction)
                                        const newFocusedCell = `${nextCell.marketplace}-${nextCell.style}-${nextCell.size}`
                                        setFocusedCell(newFocusedCell)
                                        requestAnimationFrame(() => {
                                          const nextCellButton = document.querySelector(`button[data-cell="${newFocusedCell}"]`)
                                          if (nextCellButton) {
                                            nextCellButton.focus()
                                            nextCellButton.click()
                                          }
                                        })
                                        return
                                      }
                                      
                                      if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(e.key)) {
                                        
                                        const input = e.target
                                        const cursorPosition = input.selectionStart
                                        const inputLength = input.value.length

                                        if (e.key === 'ArrowLeft' && cursorPosition === 0) {
                                          e.preventDefault()
                                          e.stopPropagation()
                                          handleKeyNavigation(e, marketplace.key, style.key, size.key)
                                        } else if (e.key === 'ArrowRight' && cursorPosition === inputLength) {
                                          e.preventDefault()
                                          e.stopPropagation()
                                          handleKeyNavigation(e, marketplace.key, style.key, size.key)
                                        } else if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
                                          e.preventDefault()
                                          e.stopPropagation()
                                          handleKeyNavigation(e, marketplace.key, style.key, size.key)
                                        }
                                      }
                                    }}
                                  />
                                  <button
                                    onClick={handleSaveCell}
                                    disabled={saving}
                                    className="p-1 text-green-600 hover:text-green-800 disabled:opacity-50"
                                    title="Save article"
                                  >
                                    <Check className="w-4 h-4" />
                                  </button>
                                  <button
                                    onClick={handleCancelEdit}
                                    disabled={saving}
                                    className="p-1 text-red-600 hover:text-red-800 disabled:opacity-50"
                                  >
                                    <X className="w-4 h-4" />
                                  </button>
                                </div>
                              ) : (
                                <div className="flex items-center space-x-1">
                                  <button
                                    onClick={() => handleCellClick(marketplace.key, style.key, size.key)}
                                    onKeyDown={(e) => handleKeyNavigation(e, marketplace.key, style.key, size.key)}
                                    onFocus={() => setFocusedCell(`${marketplace.key}-${style.key}-${size.key}`)}
                                    tabIndex={0}
                                    data-cell={`${marketplace.key}-${style.key}-${size.key}`}
                                    className={`flex-1 px-2 py-1 text-sm text-center border rounded focus:outline-none transition-all min-h-[32px] flex items-center justify-center ${isSaved
                                      ? 'border-green-500 bg-green-100 text-green-800'
                                      : focusedCell === `${marketplace.key}-${style.key}-${size.key}`
                                        ? 'border-blue-500 bg-blue-100 ring-2 ring-blue-300'
                                        : 'border-gray-200 hover:border-blue-500 hover:bg-blue-50 focus:border-blue-500 focus:bg-blue-50'
                                      }`}
                                  >
                                    {value ? (
                                      <span className={`font-mono ${isSaved ? 'text-green-800' : 'text-gray-900'}`}>{value}</span>
                                    ) : (
                                      <span className={isSaved ? 'text-green-600' : 'text-gray-400'}>—</span>
                                    )}
                                  </button>
                                  {value && (
                                    <button
                                      onClick={() => handleGenerateURL(marketplace.key, style.key, size.key)}
                                      className="p-1 text-blue-600 hover:text-blue-800 hover:bg-blue-100 rounded transition-colors"
                                      title={t('partners.generate_product_url')}
                                    >
                                      <ExternalLink className="w-3 h-3" />
                                    </button>
                                  )}
                                </div>
                              )}
                            </td>
                          )
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}

export default PartnerArticlesGrid