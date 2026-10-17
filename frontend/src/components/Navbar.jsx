import { Link, useNavigate } from 'react-router-dom';
import { Bus, Train, Ship, User, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export default function Navbar() {
    const { isAuthenticated, user, logout } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <nav style={styles.nav}>
            <div>
                <Link to="/" style={styles.logo}>
                    Ticket<span style={{ color: '#0066cc' }}>Katon</span>
                </Link>
            </div>

            <div style={styles.services}>
                <Link to="/?type=bus" style={styles.serviceItem}>
                    <Bus size={18} />
                    <span>Bus</span>
                </Link>
                <Link to="/?type=train" style={styles.serviceItem}>
                    <Train size={18} />
                    <span>Train</span>
                </Link>
                <Link to="/?type=launch" style={styles.serviceItem}>
                    <Ship size={18} />
                    <span>Launch</span>
                </Link>
            </div>

            <div>
                {isAuthenticated ? (
                    <div style={styles.userSection}>
                        <Link to="/profile" style={styles.profileLink}>
                            <User size={18} />
                            <span>{user?.name || 'Profile'}</span>
                        </Link>
                        <button onClick={handleLogout} style={styles.logoutBtn}>
                            <LogOut size={16} />
                            <span>Logout</span>
                        </button>
                    </div>
                ) : (
                    <div style={styles.guestSection}>
                        <Link to="/login" style={styles.loginBtn}>Login</Link>
                        <Link to="/register" style={styles.signupBtn}>Sign Up</Link>
                    </div>
                )}
            </div>
        </nav>
    );
}

const styles = {
    nav: {
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: '12px 24px',
        backgroundColor: '#ffffff',
        borderBottom: '1px solid #e5e7eb',
        boxShadow: '0 2px 4px rgba(0,0,0,0.05)'
    },
    logo: { fontSize: '20px', fontWeight: 'bold', textDecoration: 'none', color: '#111827' },
    services: { display: 'flex', gap: '20px' },
    serviceItem: {
        display: 'flex',
        alignItems: 'center',
        gap: '6px',
        textDecoration: 'none',
        color: '#374151',
        fontWeight: '500',
        fontSize: '15px'
    },
    userSection: { display: 'flex', alignItems: 'center', gap: '16px' },
    profileLink: { display: 'flex', alignItems: 'center', gap: '6px', textDecoration: 'none', color: '#111827' },
    logoutBtn: {
        display: 'flex',
        alignItems: 'center',
        gap: '4px',
        padding: '6px 12px',
        backgroundColor: '#fee2e2',
        color: '#b91c1c',
        border: 'none',
        borderRadius: '4px',
        cursor: 'pointer'
    },
    guestSection: { display: 'flex', gap: '12px' },
    loginBtn: {
        padding: '8px 16px',
        textDecoration: 'none',
        color: '#0066cc',
        border: '1px solid #0066cc',
        borderRadius: '4px',
        fontWeight: '500'
    },
    signupBtn: {
        padding: '8px 16px',
        textDecoration: 'none',
        backgroundColor: '#0066cc',
        color: '#ffffff',
        borderRadius: '4px',
        fontWeight: '500'
    }
};