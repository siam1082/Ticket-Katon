import { createContext, useContext, useState, useEffect } from 'react';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [token, setToken] = useState(localStorage.getItem('token') || null);
    const [role, setRole] = useState(localStorage.getItem('role') || null);

    useEffect(() => {
        const savedUser = localStorage.getItem('user');
        if (savedUser && token) {
            try {
                setUser(JSON.parse(savedUser));
            } catch (e) {
                console.error('Failed to parse user data from localStorage', e);
            }
        }
    }, [token]);

    const login = (userData, authToken, userRole = 'user') => {
        setUser(userData);
        setToken(authToken);
        setRole(userRole);

        localStorage.setItem('token', authToken);
        localStorage.setItem('role', userRole);
        localStorage.setItem('user', JSON.stringify(userData));
    };

    const logout = () => {
        setUser(null);
        setToken(null);
        setRole(null);

        localStorage.removeItem('token');
        localStorage.removeItem('role');
        localStorage.removeItem('user');
    };

    return (
        <AuthContext.Provider value={{ user, token, role, login, logout, isAuthenticated: !!token }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);