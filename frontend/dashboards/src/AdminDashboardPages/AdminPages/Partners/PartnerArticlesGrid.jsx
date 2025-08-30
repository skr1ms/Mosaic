import React, { useState, useEffect } from 'react'
import { Save, X, Edit2, Check, AlertCircle } from 'lucide-react'
import { useParams, useNavigate } from 'react-router-dom'
import { toast } from 'react-toastify'
import axios from 'axios'

const PartnerArticlesGrid = () => {
  const { partnerId } = useParams()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [articles, setArticles] = useState({})
  const [editingCell, setEditingCell] = useState(null)
  const [tempValue, setTempValue] = useState('')
  const [partner, setPartner] = useState(null)

  const sizes = [
    { key: '20x20', title: '20×20 см' },
    { key: '30x40', title: '30×40 см' },
    { key: '40x40', title: '40×40 см' },
    { key: '40x50', title: '40×50 см' },
    { key: '40x60', title: '40×60 см' },
    { key: '50x70', title: '50×70 см' }
  ]

  const styles = [
    { key: 'grayscale', title: 'Черно-белый' },
    { key: 'skin_tones', title: 'Телесные тона' },
    { key: 'pop_art', title: 'Поп-арт' },
    { key: 'max_colors', title: 'Максимум цветов' }
  ]

  const marketplaces = [
    { key: 'ozon', title: 'OZON', color: 'bg-orange-100 text-orange-800' },
    { key: 'wildberries', title: 'Wildberries', color: 'bg-purple-100 text-purple-800' }
  ]

  useEffect(() => {
    fetchPartnerData()
    fetchArticles()
  }, [partnerId])

  const fetchPartnerData = async () => {
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
      toast.error('Ошибка загрузки данных партнёра')
    }
  }

  const fetchArticles = async () => {
    try {
      const token = localStorage.getItem('adminToken')
      const response = await axios.get(
        `${import.meta.env.VITE_API_URL}/api/admin/partners/${partnerId}/articles/grid`,
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      )
      
      // Преобразуем массив в объект для удобного доступа
      const articlesMap = {}
      response.data.forEach(article => {
        const key = `${article.marketplace}-${article.style}-${article.size}`
        articlesMap[key] = article.sku || ''
      })
      
      setArticles(articlesMap)
      setLoading(false)
    } catch (error) {
      console.error('Error fetching articles:', error)
      toast.error('Ошибка загрузки артикулов')
      setLoading(false)
    }
  }

  const handleCellClick = (marketplace, style, size) => {
    const key = `${marketplace}-${style}-${size}`
    setEditingCell(key)
    setTempValue(articles[key] || '')
  }

  const handleSaveCell = async () => {
    if (!editingCell) return

    const [marketplace, style, size] = editingCell.split('-')
    
    setSaving(true)
    try {
      const token = localStorage.getItem('adminToken')
      await axios.put(
        `${import.meta.env.VITE_API_URL}/api/admin/partners/${partnerId}/articles/sku`,
        {
          size,
          style,
          marketplace,
          sku: tempValue
        },
        {
          headers: { Authorization: `Bearer ${token}` }
        }
      )
      
      setArticles(prev => ({
        ...prev,
        [editingCell]: tempValue
      }))
      
      setEditingCell(null)
      setTempValue('')
      toast.success('Артикул сохранён')
    } catch (error) {
      console.error('Error saving article:', error)
      toast.error('Ошибка сохранения артикула')
    } finally {
      setSaving(false)
    }
  }

  const handleCancelEdit = () => {
    setEditingCell(null)
    setTempValue('')
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
          <p className="mt-4 text-gray-600">Загрузка артикулов...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-6">
        <button
          onClick={() => navigate('/admin/partners')}
          className="text-blue-600 hover:text-blue-800 mb-4 inline-flex items-center"
        >
          ← Назад к партнёрам
        </button>
        
        <div className="bg-white rounded-lg shadow p-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            Артикулы товаров
          </h1>
          {partner && (
            <p className="text-gray-600">
              Партнёр: <span className="font-semibold">{partner.brand_name}</span>
            </p>
          )}
          
          <div className="mt-4 p-4 bg-blue-50 rounded-lg">
            <div className="flex items-start">
              <AlertCircle className="w-5 h-5 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
              <div className="text-sm text-blue-800">
                <p className="font-semibold mb-1">Как работает таблица артикулов:</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>Каждая ячейка представляет уникальную комбинацию: маркетплейс → стиль → размер</li>
                  <li>Нажмите на ячейку, чтобы ввести или изменить артикул товара</li>
                  <li>Артикулы используются для формирования прямых ссылок на товары в маркетплейсах</li>
                  <li>Оставьте ячейку пустой, если товар с такими параметрами отсутствует</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Articles Grid for each marketplace */}
      {marketplaces.map(marketplace => (
        <div key={marketplace.key} className="mb-8">
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className={`px-6 py-4 ${marketplace.color}`}>
              <h2 className="text-lg font-semibold">{marketplace.title}</h2>
            </div>
            
            <div className="p-6">
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead>
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Стиль / Размер
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
                          
                          return (
                            <td key={size.key} className="px-2 py-2">
                              {isEditingCell ? (
                                <div className="flex items-center space-x-1">
                                  <input
                                    type="text"
                                    value={tempValue}
                                    onChange={(e) => setTempValue(e.target.value)}
                                    className="w-full px-2 py-1 text-sm border border-blue-500 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                                    placeholder="Артикул"
                                    autoFocus
                                    onKeyDown={(e) => {
                                      if (e.key === 'Enter') handleSaveCell()
                                      if (e.key === 'Escape') handleCancelEdit()
                                    }}
                                  />
                                  <button
                                    onClick={handleSaveCell}
                                    disabled={saving}
                                    className="p-1 text-green-600 hover:text-green-800 disabled:opacity-50"
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
                                <button
                                  onClick={() => handleCellClick(marketplace.key, style.key, size.key)}
                                  className="w-full px-2 py-1 text-sm text-center border border-gray-200 rounded hover:border-blue-500 hover:bg-blue-50 transition-colors min-h-[32px] flex items-center justify-center"
                                >
                                  {value ? (
                                    <span className="font-mono text-gray-900">{value}</span>
                                  ) : (
                                    <span className="text-gray-400">—</span>
                                  )}
                                </button>
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

      {/* Templates Info */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">
          Шаблоны ссылок
        </h3>
        
        <div className="space-y-4">
          {partner?.ozon_link_template && (
            <div>
              <p className="text-sm font-medium text-gray-700 mb-1">OZON:</p>
              <code className="block p-2 bg-gray-100 rounded text-sm text-gray-800">
                {partner.ozon_link_template}
              </code>
              <p className="text-xs text-gray-500 mt-1">
                Используйте {'{sku}'} для подстановки артикула
              </p>
            </div>
          )}
          
          {partner?.wildberries_link_template && (
            <div>
              <p className="text-sm font-medium text-gray-700 mb-1">Wildberries:</p>
              <code className="block p-2 bg-gray-100 rounded text-sm text-gray-800">
                {partner.wildberries_link_template}
              </code>
              <p className="text-xs text-gray-500 mt-1">
                Используйте {'{sku}'} для подстановки артикула
              </p>
            </div>
          )}
          
          {!partner?.ozon_link_template && !partner?.wildberries_link_template && (
            <p className="text-gray-500 text-sm">
              Шаблоны ссылок не настроены. Настройте их в профиле партнёра.
            </p>
          )}
        </div>
      </div>
    </div>
  )
}

export default PartnerArticlesGrid