import { useEffect, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Ticket, Menu, X, ShieldCheck, User, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import './Navbar.css';

// Links to parts of the home page use a #section address. "match" decides which link is highlighted.
const NAV_ITEMS = [
  { label: 'Home', to: '/', match: ({ onHome, type, hash }) => onHome && !type && !hash },
  { label: 'Bus', to: '/?type=bus', match: ({ onHome, type, hash }) => onHome && type === 'bus' && !hash },
  { label: 'Train', to: '/?type=train', match: ({ onHome, type, hash }) => onHome && type === 'train' && !hash },
  { label: 'Launch', to: '/?type=launch', match: ({ onHome, type, hash }) => onHome && type === 'launch' && !hash },
  { label: 'Routes', to: '/#routes', match: ({ onHome, hash }) => onHome && hash === '#routes' },
  { label: 'About us', to: '/#team', match: ({ onHome, hash }) => onHome && hash === '#team' },
  { label: 'Contact', to: '/#contact', match: ({ hash }) => hash === '#contact' },
];

export default function Navbar() {
  const { isAuthenticated, user, role, logout } = useAuth();
  const navigate = useNavigate();
  const { pathname, search, hash } = useLocation();
  const [open, setOpen] = useState(false);

  // Close the mobile menu after every navigation.
  useEffect(() => {
    setOpen(false);
  }, [pathname, search, hash]);

  const state = { onHome: pathname === '/', type: new URLSearchParams(search).get('type'), hash };
  const displayName = user?.full_name || user?.name || 'Profile';

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="nav">
      <div className="nav-inner">
        <Link to="/" className="nav-logo" aria-label="TicketKaton home">
          <span className="nav-logo-mark" aria-hidden="true">
            <Ticket size={18} />
          </span>
          Ticket<span>Katon</span>
        </Link>

        <button
          type="button"
          className="nav-toggle"
          aria-label={open ? 'Close menu' : 'Open menu'}
          aria-expanded={open}
          aria-controls="nav-panel"
          onClick={() => setOpen((v) => !v)}
        >
          {open ? <X size={22} /> : <Menu size={22} />}
        </button>

        <div id="nav-panel" className={open ? 'nav-panel open' : 'nav-panel'}>
          <nav className="nav-links" aria-label="Main">
            {NAV_ITEMS.map((item) => (
              <Link
                key={item.label}
                to={item.to}
                className={item.match(state) ? 'nav-link active' : 'nav-link'}
                aria-current={item.match(state) ? 'page' : undefined}
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="nav-actions">
            {isAuthenticated ? (
              <>
                {role === 'admin' && (
                  <Link to="/admin" className="nav-text-link">
                    <ShieldCheck size={16} aria-hidden="true" /> Admin
                  </Link>
                )}
                <Link to="/profile" className="nav-text-link">
                  <User size={16} aria-hidden="true" /> {displayName}
                </Link>
                <button type="button" onClick={handleLogout} className="nav-logout">
                  <LogOut size={15} aria-hidden="true" /> Logout
                </button>
              </>
            ) : (
              <>
                <Link to="/login" className="btn-secondary nav-btn">Login</Link>
                <Link to="/register" className="btn-primary nav-btn">Sign Up</Link>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
