/// <reference types="vite/client" />

declare module 'react' {
  interface JSX {
    IntrinsicElements: any;
  }
}

declare module 'react-router-dom' {
  export * from 'react-router-dom';
} 