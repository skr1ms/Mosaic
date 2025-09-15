import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';


export const TransitionGroup = ({ children, ...props }) => {
  return (
    <AnimatePresence mode="wait" {...props}>
      {children}
    </AnimatePresence>
  );
};


export const CSSTransition = ({ 
  children, 
  classNames = 'TabsAnimation',
  timeout = 1500,
  appear = true,
  enter = false,
  exit = false,
  component: Component = 'div',
  className = '',
  ...props 
}) => {
  
  // For critical layout components (header, sidebar), render without animation wrappers
  // to ensure CSS rules like .fixed-header.fixed-sidebar .app-sidebar .app-header__logo work properly
  if (className && (className.includes('app-header') || className.includes('app-sidebar'))) {
    
    if (Component === 'div') {
      return React.createElement('div', {
        className,
        ...props
      }, children);
    }
    return React.createElement(Component, {
      className,
      ...props
    }, children);
  }

  
  return (
    <motion.div
      key={`transition-${classNames}`}
      className={className}
      initial={appear ? { opacity: 0 } : false}
      animate={{ opacity: 1 }}
      exit={exit ? { opacity: 0 } : false}
      transition={{ duration: timeout / 1000 }}
      {...props}
    >
      {children}
    </motion.div>
  );
};

const TransitionComponents = { TransitionGroup, CSSTransition };
export default TransitionComponents; 