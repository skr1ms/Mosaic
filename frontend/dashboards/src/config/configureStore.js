import { configureStore } from "@reduxjs/toolkit";
import reducers from "../reducers";

export default function configureAppStore() {
  return configureStore({
    reducer: reducers,
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware({
        serializableCheck: {
          // Ignore these action types
          ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
        },
      }),
    // В development используем переменную из Docker Compose, в production - false
    devTools: process.env.NODE_ENV !== 'production',
  });
}
