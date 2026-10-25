import { Link } from 'react-router-dom';
import { Info, Mail } from 'lucide-react';
import { DISCLAIMER_EN, DISCLAIMER_BN } from '../constants/disclaimer';
import { TEAM } from '../data/team';
import './Footer.css';

export default function Footer() {
  const contacts = TEAM.filter((m) => m.links?.email?.trim());

  return (
    <footer className="foot">
      <div className="foot-inner">
        <div className="foot-brand">
          <h3>
            Ticket<span>Katon</span>
          </h3>
          <p>An academic prototype that shows how bus, train and launch ticket booking can work in Bangladesh.</p>
        </div>

        <div className="foot-col">
          <h4>Explore</h4>
          <ul>
            <li><Link to="/">Home</Link></li>
            <li><Link to="/?type=bus">Bus tickets</Link></li>
            <li><Link to="/?type=train">Train tickets</Link></li>
            <li><Link to="/?type=launch">Launch tickets</Link></li>
          </ul>
        </div>

        <div className="foot-col">
          <h4>Project</h4>
          <ul>
            <li><Link to="/#routes">Popular routes</Link></li>
            <li><Link to="/#team">About us</Link></li>
            <li><Link to="/login">Login</Link></li>
            <li><Link to="/register">Sign up</Link></li>
          </ul>
        </div>

        <div className="foot-col" id="contact">
          <h4>Contact</h4>
          {contacts.length > 0 ? (
            <ul>
              {contacts.map((m) => (
                <li key={m.id}>
                  <a href={`mailto:${m.links.email.trim()}`} className="foot-mail">
                    <Mail size={14} aria-hidden="true" /> {m.name}
                  </a>
                </li>
              ))}
            </ul>
          ) : (
            <p className="foot-note">
              Reach the team through the <Link to="/#team">About us</Link> section.
            </p>
          )}
        </div>
      </div>

      <div className="foot-disclaimer">
        <Info size={16} aria-hidden="true" />
        <div>
          <p>{DISCLAIMER_EN}</p>
          <p lang="bn">{DISCLAIMER_BN}</p>
        </div>
      </div>

      <div className="foot-bottom">
        © 2026 TicketKaton · Academic project by {TEAM.map((m) => m.name).join(' and ')}
      </div>
    </footer>
  );
}
