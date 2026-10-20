import { useState } from 'react';
import { Info, ChevronDown } from 'lucide-react';
import { DISCLAIMER_SHORT, DISCLAIMER_EN, DISCLAIMER_BN } from '../constants/disclaimer';
import './DisclaimerBanner.css';

export default function DisclaimerBanner() {
  const [open, setOpen] = useState(false);

  return (
    <div className="demo-banner" role="note">
      <div className="demo-banner-row">
        <Info size={16} aria-hidden="true" />
        <span className="demo-banner-text">
          <strong>Academic demo.</strong> {DISCLAIMER_SHORT}
        </span>
        <button
          type="button"
          className="demo-banner-toggle"
          aria-expanded={open}
          onClick={() => setOpen((v) => !v)}
        >
          {open ? 'Hide' : 'Details'}
          <ChevronDown size={14} className={open ? 'flip' : ''} aria-hidden="true" />
        </button>
      </div>

      {open && (
        <div className="demo-banner-more">
          <p>{DISCLAIMER_EN}</p>
          <p lang="bn">{DISCLAIMER_BN}</p>
        </div>
      )}
    </div>
  );
}
