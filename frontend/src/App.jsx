import { lazy, Suspense } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import DisclaimerBanner from './components/DisclaimerBanner';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import PageLoader from './components/PageLoader';
import Home from './pages/Home';
const Login = lazy(() => import('./pages/Login'));
const Register = lazy(() => import('./pages/Register'));
const VerifyEmail = lazy(() => import('./pages/VerifyEmail'));
const Profile = lazy(() => import('./pages/Profile'));
const SearchResults = lazy(() => import('./pages/SearchResults'));
const SeatSelect = lazy(() => import('./pages/SeatSelect'));
const AdminDashboard = lazy(() => import('./pages/AdminDashboard'));

function Padded({ children }) {
  return <div className="page-pad">{children}</div>;
}

export default function App() {
  return (
    <AuthProvider>
      <Router>
        <DisclaimerBanner />
        <Navbar />
        <main className="app-main">
          <Suspense fallback={<PageLoader fullscreen label="Loading…" />}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/verify-email" element={<VerifyEmail />} />
            <Route path="/search" element={<Padded><SearchResults /></Padded>} />
            <Route path="/seat-select" element={<SeatSelect />} />
            <Route path="/profile" element={<Padded><Profile /></Padded>} />
            <Route path="/admin" element={<Padded><AdminDashboard /></Padded>} />
          </Routes>
          </Suspense>
        </main>
        <Footer />
      </Router>
    </AuthProvider>
  );
}
