import { motion } from 'framer-motion'

export default function Button({ children, variant = 'primary', style = {}, ...props }) {
  const styles = {
    primary: {
      background: 'linear-gradient(135deg, #2563eb, #1d4ed8)',
      boxShadow: '4px 4px 10px #060f1c, -2px -2px 6px #0e1f3a',
      border: '1px solid rgba(56, 189, 248, 0.15)',
      borderRadius: 10, padding: '10px 24px',
      color: '#fff', fontSize: 14, fontWeight: 700,
      cursor: 'pointer',
    },
    secondary: {
      background: 'transparent',
      border: '1px solid rgba(37, 99, 235, 0.3)',
      borderRadius: 10, padding: '10px 24px',
      color: '#94a3b8', fontSize: 14, fontWeight: 500,
      cursor: 'pointer',
    },
    ghost: {
      background: 'none', border: 'none',
      color: '#94a3b8', fontSize: 14, cursor: 'pointer',
      padding: '8px 16px',
    },
  }

  return (
    <motion.button
      style={{ ...styles[variant], ...style }}
      whileHover={{ scale: 1.03, filter: 'brightness(1.1)' }}
      whileTap={{ scale: 0.97 }}
      {...props}
    >
      {children}
    </motion.button>
  )
}
