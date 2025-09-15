import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export const usePartnerStore = create(
  devtools(
    (set, get) => ({
      partner: null,
      setPartner: partner => set({ partner }),

      getBrandColors: () => {
        const state = get();
        return state.partner?.brandColors || [];
      },

      getPrimaryColor: () => {
        const state = get();
        return state.partner?.brandColors?.[0] || '#3B82F6';
      },

      getSecondaryColor: () => {
        const state = get();
        return state.partner?.brandColors?.[1] || '#10B981';
      },

      getAccentColor: () => {
        const state = get();
        return state.partner?.brandColors?.[2] || '#F59E0B';
      },

      editorState: {
        currentStep: 1,
        imageData: null,
        selectedOptions: null,
        schemaData: null,
      },

      setEditorState: updates =>
        set(state => ({
          editorState: { ...state.editorState, ...updates },
        })),

      resetEditor: () =>
        set(state => ({
          editorState: {
            currentStep: 1,
            imageData: null,
            selectedOptions: null,
            schemaData: null,
          },
        })),
    }),
    {
      name: 'partner-store',
    }
  )
);

export const useUIStore = create(
  devtools(
    (set, get) => ({
      notifications: [],
      nextNotificationId: 1,

      addNotification: notification => {
        const now = Date.now();
        const state = get();
        const id = state.nextNotificationId;
        const type = notification.type || 'info';
        const title = notification.title || '';
        const message = notification.message || '';
        const dedupeKey = `${type}::${title}::${message}`;

        const existingNotification = state.notifications.find(
          n => n.dedupeKey === dedupeKey
        );
        if (existingNotification) {
          return;
        }

        const newNotification = {
          id,
          ...notification,
          timestamp: now,
          dedupeKey,
        };

        set({
          notifications: [...state.notifications, newNotification],
          nextNotificationId: id + 1,
        });

        const duration =
          typeof notification.duration === 'number'
            ? notification.duration
            : 5000;
        setTimeout(() => {
          get().removeNotification(id);
        }, duration);
      },

      removeNotification: id => {
        set(state => ({
          notifications: state.notifications.filter(n => n.id !== id),
        }));
      },

      clearNotifications: () => set({ notifications: [] }),
    }),
    {
      name: 'ui-store',
    }
  )
);

export const useCouponStore = create(
  devtools(
    (set, get) => ({
      coupon: null,
      setCoupon: coupon => set({ coupon }),
      clearCoupon: () => set({ coupon: null }),
    }),
    {
      name: 'coupon-store',
    }
  )
);
