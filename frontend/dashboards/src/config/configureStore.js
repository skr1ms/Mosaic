import { configureStore } from "@reduxjs/toolkit";
import reducers from "../reducers";

export default function configureAppStore() {
  return configureStore({
    reducer: reducers,
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware({
        serializableCheck: {
          
          ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
        },
      }),
    
    devTools: process.env.NODE_ENV !== 'production',
  });
}
