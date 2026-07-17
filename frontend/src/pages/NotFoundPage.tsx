import { Link } from 'react-router-dom';

export default function NotFoundPage() {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '80vh',
      gap: '1rem',
      padding: '2rem',
      textAlign: 'center',
    }}>
      <h1 style={{ fontSize: '4rem', margin: 0, color: '#6366f1' }}>404</h1>
      <p style={{ fontSize: '1.25rem', color: '#64748b', margin: 0 }}>
        Page not found
      </p>
      <Link
        to="/"
        style={{
          marginTop: '1rem',
          padding: '0.75rem 1.5rem',
          borderRadius: '12px',
          background: '#6366f1',
          color: 'white',
          textDecoration: 'none',
          fontWeight: 600,
        }}
      >
        Back to Home
      </Link>
    </div>
  );
}
