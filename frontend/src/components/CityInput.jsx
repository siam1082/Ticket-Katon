import { useEffect, useRef, useState } from 'react';
import API from '../services/api';
import './CityInput.css';

// A city text box with suggestions from the server. It understands other spellings
// (Chittagong for Chattogram) and small typos (Rajshai for Rajshahi).
export default function CityInput({ id, label, icon, value, onChange, placeholder, required }) {
  const [items, setItems] = useState([]);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(-1);
  const boxRef = useRef(null);

  // Fetch suggestions shortly after the user stops typing.
  useEffect(() => {
    if (!open) return undefined;
    const controller = new AbortController();
    const timer = setTimeout(async () => {
      try {
        const res = await API.get('/cities', { params: { q: value, limit: 6 }, signal: controller.signal });
        setItems(res.data || []);
        setActive(-1);
      } catch {
        // Offline, or the request was cancelled by newer typing: keep what we have.
      }
    }, 150);
    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [value, open]);

  // Close when clicking elsewhere.
  useEffect(() => {
    const onDown = (e) => {
      if (boxRef.current && !boxRef.current.contains(e.target)) setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, []);

  const pick = (name) => {
    onChange(name);
    setOpen(false);
  };

  const onKeyDown = (e) => {
    if (!open || items.length === 0) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActive((a) => (a + 1) % items.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActive((a) => (a <= 0 ? items.length - 1 : a - 1));
    } else if (e.key === 'Enter' && active >= 0) {
      e.preventDefault();
      pick(items[active].name);
    } else if (e.key === 'Escape') {
      setOpen(false);
    }
  };

  const listId = `${id}-list`;

  return (
    <div className="field ci" ref={boxRef}>
      <label htmlFor={id}>{icon} {label}</label>
      <input
        id={id}
        type="text"
        role="combobox"
        aria-expanded={open && items.length > 0}
        aria-controls={listId}
        aria-autocomplete="list"
        autoComplete="off"
        value={value}
        placeholder={placeholder}
        required={required}
        maxLength={60}
        onChange={(e) => {
          onChange(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onKeyDown={onKeyDown}
      />
      {open && items.length > 0 && (
        <ul id={listId} role="listbox" className="ci-list">
          {items.map((c, i) => (
            <li
              key={c.name}
              role="option"
              aria-selected={i === active}
              className={i === active ? 'ci-item active' : 'ci-item'}
              onMouseDown={(e) => {
                e.preventDefault();
                pick(c.name);
              }}
            >
              <span>{c.name}</span>
              {c.match === 'similar' && <small>Did you mean?</small>}
              {c.match === 'alias' && <small>Also spelled this way</small>}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
