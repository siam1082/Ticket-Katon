import { Link, useNavigate } from 'react-router-dom';
import { Bus, Train, Ship, User, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import './Navbar.css';

export default function Navbar() {
    const { isAuthenticated, user, logout } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <nav className="gov-navbar">
            <div>
                <Link to="/" className="gov-logo">
                    Ticket<span className="gov-accent">Katon</span>
                </Link>
            </div>

            <div className="gov-services">
                <Link to="/?type=bus" className="gov-service-item">
                    <Bus size={18} />
                    <span>Bus</span>
                </Link>
                <Link to="/?type=train" className="gov-service-item">
                    <Train size={18} />
                    <span>Train</span>
                </Link>
                <Link to="/?type=launch" className="gov-service-item">
                    <Ship size={18} />
                    <span>Launch</span>
                </Link>
            </div>

            <div>
                {isAuthenticated ? (
                    <div className="gov-user-section">
                        <Link to="/profile" className="gov-profile-link">
                            <User size={18} />
                            <span>{user?.name || 'Profile'}</span>
                        </Link>
                        <button onClick={handleLogout} className="gov-logout-btn">
                            <LogOut size={16} />
                            <span>Logout</span>
                        </button>
                    </div>
                ) : (
                    <div className="gov-guest-section">
                        <Link to="/login" className="gov-login-btn">Login</Link>
                        <Link to="/register" className="gov-signup-btn">Sign Up</Link>
                    </div>
                )}
            </div>
        </nav>
    );
}
