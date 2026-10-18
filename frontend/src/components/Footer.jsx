import { Link } from 'react-router-dom';
import { Facebook, Instagram, Mail, MapPin, Phone } from 'lucide-react';
import './Footer.css';

export default function Footer() {
  return (
    <footer className="gov-footer">
      <div className="gov-footer-top">
        <div className="gov-footer-brand">
          <h3>Ticket<span>Katon</span></h3>
          <p>Government transportation and ticketing services for safe, reliable travel across Bangladesh.</p>
        </div>

        <div className="gov-footer-links">
          <h4>Quick Links</h4>
          <ul>
            <li><Link to="/">Home</Link></li>
            <li><Link to="/?type=bus">Bus Tickets</Link></li>
            <li><Link to="/?type=train">Train Tickets</Link></li>
            <li><Link to="/?type=launch">Launch Tickets</Link></li>
          </ul>
        </div>

        <div className="gov-footer-links">
          <h4>Support</h4>
          <ul>
            <li><a href="#">Help Center</a></li>
            <li><a href="#">Ticket Policies</a></li>
            <li><a href="#">Contact Desk</a></li>
            <li><a href="#">Customer Care</a></li>
          </ul>
        </div>

        <div className="gov-footer-contact">
          <h4>Contact</h4>
          <ul>
            <li><Phone size={14} /> +880 1700-000000</li>
            <li><Mail size={14} /> support@ticketkaton.gov.bd</li>
            <li><MapPin size={14} /> Dhaka, Bangladesh</li>
          </ul>
        </div>
      </div>

      <div className="gov-footer-bottom">
        <span>© 2026 TicketKaton. All rights reserved.</span>
        <div className="gov-socials">
          <a href="#" aria-label="Facebook"><Facebook size={16} /></a>
          <a href="#" aria-label="Instagram"><Instagram size={16} /></a>
        </div>
      </div>
    </footer>
  );
}
