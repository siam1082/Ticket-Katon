import { Link, NavLink, useNavigate } from 'react-router-dom';
import { Bus, Train, Ship, User, LogOut, Ticket, ShieldCheck } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import './Navbar.css';

const SERVICES = [
  { type: 'bus', label: 'Bus', Icon: Bus },
  { type: 'train', label: 'Train', Icon: Train },
  { type: 'launch', label: 'Launch', Icon: Ship },
];

export default function Navbar() {
  const { isAuthenticated, user, role, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const displayName = user?.full_name || user?.name || 'Profile';

  return (
    <header className="nav">
      <div className="nav-inner">
        <Link to="/" className="nav-logo" aria-label="TicketKaton home">
          <span className="nav-logo-mark" aria-hidden="true">
            <Ticket size={18} />
          </span>
          Ticket<span>Katon</span>
        </Link>

        <nav className="nav-services" aria-label="Transport types">
          {SERVICES.map(({ type, label, Icon }) => (
            <NavLink key={type} to={`/?type=${type}`} className="nav-service">
              <Icon size={18} aria-hidden="true" />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>

        {isAuthenticated ? (
          <div className="nav-actions">
            {role === 'admin' && (
              <Link to="/admin" className="nav-profile">
                <ShieldCheck size={18} aria-hidden="true" />
                <span>Admin</span>
              </Link>
            )}
            <Link to="/profile" className="nav-profile">
              <User size={18} aria-hidden="true" />
              <span>{displayName}</span>
            </Link>
            <button type="button" onClick={handleLogout} className="nav-logout">
              <LogOut size={16} aria-hidden="true" />
              <span>Logout</span>
            </button>
          </div>
        ) : (
          <div className="nav-actions">
            <Link to="/login" className="btn-secondary nav-btn">Login</Link>
            <Link to="/register" className="btn-primary nav-btn">Sign Up</Link>
          </div>
        )}
      </div>
    </header>
  );
}
