import React from 'react';
import { ArrowPathIcon } from '@heroicons/react/24/outline';
import { motion, AnimatePresence } from 'framer-motion';

const LoadingSpinner = ({
  text = 'Loading...',
  size = 72,
  color = 'text-black-600',
  fullscreen = false,
  visible = true
}) => {
  return (
    <AnimatePresence>
      {visible && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className={`flex flex-col items-center justify-center ${fullscreen ? 'h-screen w-screen' : ''
            }`}
        >
          <ArrowPathIcon className={`h-${size} w-${size} animate-spin ${color}`} />
          <p className="mt-2 text-gray-600 dark:text-gray-300">{text}</p>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default LoadingSpinner;
