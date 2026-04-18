export default function Input({ label, style = {}, ...props }) {
  return (
    <div style={{ marginBottom: 16 }}>
      {label && (
        <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: '#94a3b8', marginBottom: 8 }}>
          {label}
        </label>
      )}
      <input
        style={{
          width: '100%',
          background: 'linear-gradient(145deg, #060f1c, #0f2140)',
          boxShadow: 'inset 3px 3px 6px #050d18, inset -3px -3px 6px #0e1f3a',
          borderRadius: 10,
          border: '1px solid rgba(37, 99, 235, 0.12)',
          padding: '12px 16px',
          color: '#f8fafc',
          fontSize: 14,
          outline: 'none',
          transition: 'border-color 0.3s',
          ...style,
        }}
        onFocus={(e) => e.target.style.borderColor = '#2563eb'}
        onBlur={(e) => e.target.style.borderColor = 'rgba(37, 99, 235, 0.12)'}
        {...props}
      />
    </div>
  )
}
