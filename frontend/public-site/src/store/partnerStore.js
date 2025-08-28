import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

// Store для управления партнером и брендингом
export const usePartnerStore = create(
  devtools(
    (set, get) => ({
      partner: null,
      setPartner: (partner) => set({ partner }),
      
      // Получение цветов брендинга
      getBrandColors: () => {
        const state = get()
        return state.partner?.brandColors || []
      },
      
      // Получение основного цвета (первый из массива)
      getPrimaryColor: () => {
        const state = get()
        return state.partner?.brandColors?.[0] || '#3B82F6' // fallback к синему
      },
      
      // Получение вторичного цвета (второй из массива)
      getSecondaryColor: () => {
        const state = get()
        return state.partner?.brandColors?.[1] || '#10B981' // fallback к зеленому
      },
      
      // Получение акцентного цвета (третий из массива)
      getAccentColor: () => {
        const state = get()
        return state.partner?.brandColors?.[2] || '#F59E0B' // fallback к оранжевому
      },
      
      // Состояние для редактора
      editorState: {
        currentStep: 1,
        imageData: null,
        selectedOptions: null,
        schemaData: null
      },
      
      setEditorState: (updates) => set((state) => ({
        editorState: { ...state.editorState, ...updates }
      })),
      
      resetEditor: () => set((state) => ({
        editorState: {
          currentStep: 1,
          imageData: null,
          selectedOptions: null,
          schemaData: null
        }
      }))
    }),
    {
      name: 'partner-store'
    }
  )
)

// Store для управления UI состоянием
export const useUIStore = create(
  devtools(
    (set, get) => ({
      notifications: [],
      nextNotificationId: 1,
      
      addNotification: (notification) => {
        const now = Date.now()
        const state = get()
        const id = state.nextNotificationId
        // Формируем ключ для дедупликации
        const type = notification.type || 'info'
        const title = notification.title || ''
        const message = notification.message || ''
        const dedupeKey = `${type}::${title}::${message}`

        // Проверяем, есть ли уже такое же уведомление
        const existingNotification = state.notifications.find(n => n.dedupeKey === dedupeKey)
        if (existingNotification) {
          // Если уведомление уже есть, обновляем его timestamp и не добавляем дубликат
          return
        }

        const newNotification = {
          id,
          ...notification,
          timestamp: now,
          dedupeKey,
        }

        set({
          notifications: [...state.notifications, newNotification],
          nextNotificationId: id + 1,
        })

        // Автоматически удаляем уведомление через 5 секунд (или кастомный duration)
        const duration = typeof notification.duration === 'number' ? notification.duration : 5000
        setTimeout(() => {
          get().removeNotification(id)
        }, duration)
      },
      
      removeNotification: (id) => {
        set((state) => ({
          notifications: state.notifications.filter(n => n.id !== id)
        }))
      },
      
      clearNotifications: () => set({ notifications: [] })
    }),
    {
      name: 'ui-store'
    }
  )
)

// Store для управления купоном
export const useCouponStore = create(
  devtools(
    (set, get) => ({
      coupon: null,
      setCoupon: (coupon) => set({ coupon }),
      clearCoupon: () => set({ coupon: null })
    }),
    {
      name: 'coupon-store'
    }
  )
)
