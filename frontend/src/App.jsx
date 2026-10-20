import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import DisclaimerBanner from './components/DisclaimerBanner';
import Navbar from './components/Navbar';
import Footer from './components/Footer';
import Home from './pages/Home';
import Login from './pages/Login';
import Register from './pages/Register';
import Profile from './pages/Profile';
import SearchResults from './pages/SearchResults';
import AdminDashboard from './pages/AdminDashboard';

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
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/search" element={<Padded><SearchResults /></Padded>} />
            <Route path="/profile" element={<Padded><Profile /></Padded>} />
            <Route path="/admin" element={<Padded><AdminDashboard /></Padded>} />
          </Routes>
        </main>
        <Footer />
      </Router>
    </AuthProvider>
  );
}
