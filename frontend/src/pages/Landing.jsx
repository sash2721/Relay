import { motion } from 'framer-motion'
import { useNavigate } from 'react-router-dom'
import Navbar from '../components/Navbar'
import Card from '../components/Card'
import Button from '../components/Button'

const techStack = [
  'React', 'Vue', 'Angular', 'Next.js', 'Svelte', 'Node.js', 'Go', 'Python', 'Java', 'Static HTML',
  'React', 'Vue', 'Angular', 'Next.js', 'Svelte', 'Node.js', 'Go', 'Python', 'Java', 'Static HTML',
]

const features = [
  { icon: '🚀', title: 'One-Click Deploy', desc: 'Paste your GitHub URL. We handle the rest.' },
  { icon: '🔍', title: 'Auto-Detect', desc: 'Relay figures out your stack automatically.' },
  { icon: '📡', title: 'Live Build Logs', desc: 'Watch your build happen in real-time via SSE.' },
  { icon: '🐳', title: 'Isolated Builds', desc: 'Every build runs in its own Docker container.' },
  { icon: '🌐', title: 'Instant URLs', desc: 'Your app gets a unique subdomain instantly.' },
  { icon: '⚡', title: 'Blazing Fast', desc: 'From git push to live in under 200 seconds.' },
]

const steps = [
  { num: '01', title: 'Paste Repo URL', desc: 'Drop your GitHub repository link' },
  { num: '02', title: 'We Detect & Build', desc: 'Auto-detect stack, build in Docker' },
  { num: '03', title: 'Go Live', desc: 'Your site is live at a unique URL' },
]

export default function Landing() {
  const navigate = useNavigate()

  return (
    <div style={{
      minHeight: '100vh',
      backgroundImage: 'linear-gradient(rgba(37,99,235,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(37,99,235,0.04) 1px, transparent 1px)',
      backgroundSize: '60px 60px',
    }}>
      <Navbar />

      {/* Hero */}
      <section style={{ minHeight: '90vh', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: '80px 24px 40px', position: 'relative' }}>
        {/* Radial glow */}
        <div style={{
          position: 'absolute', top: '40%', left: '50%', transform: 'translate(-50%, -50%)',
          width: 700, height: 700,
          background: 'radial-gradient(circle, rgba(37,99,235,0.12), transparent 70%)',
          pointerEvents: 'none',
        }} />

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7 }}
          style={{ textAlign: 'center', position: 'relative', zIndex: 1 }}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.2 }}
            style={{
              display: 'inline-block', marginBottom: 24,
              padding: '6px 16px', borderRadius: 20,
              background: 'rgba(20, 184, 166, 0.08)',
              border: '1px solid rgba(20, 184, 166, 0.25)',
              color: '#14b8a6', fontSize: 12, fontWeight: 600, letterSpacing: 1, textTransform: 'uppercase',
            }}
          >
            🚀 Deploy anything in seconds
          </motion.div>

          <h1 style={{ fontSize: 'clamp(48px, 8vw, 88px)', fontWeight: 900, lineHeight: 1.05, marginBottom: 24 }}>
            <span style={{ display: 'block', background: 'linear-gradient(to right, #f8fafc, #38bdf8, #14b8a6)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
              From Localhost
            </span>
            <span style={{ display: 'block', background: 'linear-gradient(to right, #facc15, #38bdf8, #2563eb)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent' }}>
              to Liftoff
            </span>
          </h1>

          <motion.p
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
            style={{ fontSize: 18, color: '#94a3b8', maxWidth: 560, margin: '0 auto 40px', lineHeight: 1.7 }}
          >
            Relay takes your GitHub repo, detects the stack, builds it in an isolated container, and serves it at a unique URL.
          </motion.p>

          <motion.div
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
            style={{ display: 'flex', gap: 16, justifyContent: 'center', flexWrap: 'wrap' }}
          >
            <Button onClick={() => navigate('/signup')}>Start Deploying →</Button>
            <Button variant="secondary">See How It Works</Button>
          </motion.div>
        </motion.div>

        {/* Floating cards */}
        <motion.div
          animate={{ y: [0, -12, 0] }}
          transition={{ duration: 4, repeat: Infinity, ease: 'easeInOut' }}
          style={{ position: 'absolute', bottom: 120, right: 60 }}
        >
          <Card style={{ padding: 20, textAlign: 'center', display: 'none' }} className="hidden lg:block">
            <div style={{ fontSize: 36, fontWeight: 900, color: '#facc15' }}>200</div>
            <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 4 }}>Seconds to Live</div>
          </Card>
        </motion.div>
      </section>

      {/* Tech Marquee */}
      <div style={{
        transform: 'rotate(-1deg)',
        background: '#facc15',
        overflow: 'hidden',
        margin: '0 -20px',
      }}>
        <motion.div
          animate={{ x: ['0%', '-50%'] }}
          transition={{ duration: 20, repeat: Infinity, ease: 'linear' }}
          style={{ display: 'flex', whiteSpace: 'nowrap', padding: '14px 0' }}
        >
          {techStack.map((tech, i) => (
            <span key={i} style={{ margin: '0 32px', fontSize: 16, fontWeight: 900, color: '#000', letterSpacing: 1 }}>
              {tech} •
            </span>
          ))}
        </motion.div>
      </div>

      {/* How It Works */}
      <section style={{ maxWidth: 1200, margin: '0 auto', padding: '120px 24px' }}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          style={{ textAlign: 'center', marginBottom: 64 }}
        >
          <h2 style={{ fontSize: 42, fontWeight: 900, marginBottom: 12 }}>
            Three Steps. <span style={{ color: '#14b8a6' }}>That's It.</span>
          </h2>
          <p style={{ color: '#94a3b8', fontSize: 16 }}>No YAML. No Dockerfiles. No CI/CD pipelines.</p>
        </motion.div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', gap: 24 }}>
          {steps.map((step, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.15 }}
            >
              <Card hover style={{ padding: 32, position: 'relative', overflow: 'hidden', cursor: 'default' }}>
                <div style={{ position: 'absolute', top: -8, right: 8, fontSize: 72, fontWeight: 900, color: 'rgba(37,99,235,0.06)' }}>
                  {step.num}
                </div>
                <div style={{ position: 'relative', zIndex: 1 }}>
                  <div style={{ fontSize: 12, fontWeight: 700, color: '#2563eb', marginBottom: 12, letterSpacing: 1.5, textTransform: 'uppercase' }}>
                    Step {step.num}
                  </div>
                  <h3 style={{ fontSize: 22, fontWeight: 700, marginBottom: 8 }}>{step.title}</h3>
                  <p style={{ color: '#94a3b8', fontSize: 14, lineHeight: 1.6 }}>{step.desc}</p>
                </div>
              </Card>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Features */}
      <section style={{ maxWidth: 1200, margin: '0 auto', padding: '0 24px 120px' }}>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          style={{ textAlign: 'center', marginBottom: 64 }}
        >
          <h2 style={{ fontSize: 42, fontWeight: 900, marginBottom: 12 }}>
            Built for <span style={{ color: '#facc15' }}>Speed</span>
          </h2>
          <p style={{ color: '#94a3b8', fontSize: 16 }}>Everything you need. Nothing you don't.</p>
        </motion.div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 20 }}>
          {features.map((f, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08 }}
            >
              <Card hover style={{ padding: 28, cursor: 'default' }}>
                <div style={{ fontSize: 32, marginBottom: 12 }}>{f.icon}</div>
                <h3 style={{ fontSize: 18, fontWeight: 700, marginBottom: 8 }}>{f.title}</h3>
                <p style={{ color: '#94a3b8', fontSize: 14, lineHeight: 1.6 }}>{f.desc}</p>
              </Card>
            </motion.div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section style={{ maxWidth: 800, margin: '0 auto', padding: '0 24px 120px', textAlign: 'center' }}>
        <Card style={{ padding: 64, position: 'relative', overflow: 'hidden' }}>
          <div style={{ position: 'absolute', inset: 0, background: 'linear-gradient(135deg, rgba(37,99,235,0.08), transparent)', pointerEvents: 'none' }} />
          <div style={{ position: 'relative', zIndex: 1 }}>
            <h2 style={{ fontSize: 42, fontWeight: 900, marginBottom: 16 }}>
              Ready to <span style={{ color: '#14b8a6' }}>Launch</span>?
            </h2>
            <p style={{ color: '#94a3b8', fontSize: 16, marginBottom: 32, maxWidth: 400, margin: '0 auto 32px' }}>
              Your next deployment is one click away. No credit card. No setup. Just ship.
            </p>
            <Button onClick={() => navigate('/signup')} style={{ fontSize: 16, padding: '14px 32px' }}>
              Get Started Free 🚀
            </Button>
          </div>
        </Card>
      </section>

      {/* Footer */}
      <footer style={{ borderTop: '1px solid rgba(37,99,235,0.1)', padding: '24px 0' }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', padding: '0 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ fontFamily: 'Lobster, cursive', color: '#38bdf8', fontSize: 20 }}>Relay</span>
          <span style={{ fontSize: 13, color: '#94a3b8' }}>Built with 🤍 and Go</span>
        </div>
      </footer>
    </div>
  )
}
