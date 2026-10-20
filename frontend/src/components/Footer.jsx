import { Link } from 'react-router-dom';
import { Info } from 'lucide-react';
import { DISCLAIMER_EN, DISCLAIMER_BN } from '../constants/disclaimer';
import './Footer.css';

export default function Footer() {
  return (
    <footer className="foot">
      <div className="foot-inner">
        <div className="foot-brand">
          <h3>
            Ticket<span>Katon</span>
          </h3>
          <p>An academic prototype showing how bus, train and launch ticket booking can work in Bangladesh.</p>
        </div>

        <div className="foot-col">
          <h4>Explore</h4>
          <ul>
            <li><Link to="/">Home</Link></li>
            <li><Link to="/?type=bus">Bus</Link></li>
            <li><Link to="/?type=train">Train</Link></li>
            <li><Link to="/?type=launch">Launch</Link></li>
          </ul>
        </div>

        <div className="foot-col">
          <h4>Account</h4>
          <ul>
            <li><Link to="/login">Login</Link></li>
            <li><Link to="/register">Sign Up</Link></li>
            <li><Link to="/profile">My bookings</Link></li>
          </ul>
        </div>
      </div>

      <div className="foot-disclaimer">
        <Info size={16} aria-hidden="true" />
        <div>
          <p>{DISCLAIMER_EN}</p>
          <p lang="bn">{DISCLAIMER_BN}</p>
        </div>
      </div>

      <div className="foot-bottom">© 2026 TicketKaton · Academic project</div>
    </footer>
  );
}
