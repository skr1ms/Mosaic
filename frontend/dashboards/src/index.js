import React from "react";
import { createRoot } from 'react-dom/client';
import "./i18n/config";

import { HashRouter } from "react-router-dom";
import "./assets/base.scss";
import Main from "./AdminDashboardPages/Main";
import ResetPassword from "./pages/ResetPassword";
import Login from "./pages/Login";
import { Routes, Route } from 'react-router-dom';
import configureAppStore from "./config/configureStore";
import { Provider } from "react-redux";

const store = configureAppStore();
const rootElement = document.getElementById("root");

const renderApp = (Component) => (
  <React.StrictMode>
    <Provider store={store}>
      <HashRouter
        future={{
          v7_startTransition: true,
          v7_relativeSplatPath: true
        }}
      >
        <Routes>
          <Route path="/reset" element={<ResetPassword />} />
          <Route path="/login" element={<Login />} />
          <Route path="/*" element={<Component />} />
        </Routes>
      </HashRouter>
    </Provider>
  </React.StrictMode>
);

const root = createRoot(rootElement);
root.render(renderApp(Main));

if (module.hot) {
  module.hot.accept("./AdminDashboardPages/Main", () => {
    const NextApp = require("./AdminDashboardPages/Main").default;
    root.render(renderApp(NextApp));
  });
}