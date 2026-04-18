import { motion } from 'framer-motion'

export default function Card({ children, style = {}, hover = false, onClick, ...props }) {
  const baseStyle = {
    background: 'linear-gradient(145deg, #0f2140, #0b1a30)',
    boxShadow: '6px 6px 14px #060f1c, -6px -6px 14px #0e1f3a',
    borderRadius: 16,
    border: '1px solid rgba(37, 99, 235, 0.1)',
    ...style,
  }

  if (hover) {
    return (
      <motion.div
        style={baseStyle}
        whileHover={{
          boxShadow: '8px 8px 20px #060f1c, -8px -8px 20px #0e1f3a, 0 0 25px rgba(37, 99, 235, 0.12)',
          borderColor: 'rgba(37, 99, 235, 0.25)',
          y: -3,
        }}
        transition={{ duration: 0.3 }}
        onClick={onClick}
        {...props}
      >
        {children}
      </motion.div>
    )
  }

  return <div style={baseStyle} {...props}>{children}</div>
}
