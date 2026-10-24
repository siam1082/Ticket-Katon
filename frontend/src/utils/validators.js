// Checks used by the sign-up and sign-in forms. The server checks the same rules again.

export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/;

// Accepts 01XXXXXXXXX, +8801XXXXXXXXX and 8801XXXXXXXXX; returns 01XXXXXXXXX or '' if invalid.
export function normalizePhone(value) {
  let v = value.replace(/[\s\-()]/g, '');
  if (v.startsWith('+88')) v = v.slice(3);
  else if (v.startsWith('88') && v.length === 13) v = v.slice(2);
  return /^01[3-9]\d{8}$/.test(v) ? v : '';
}

export function passwordChecks(password) {
  return [
    { ok: password.length >= 8, label: 'At least 8 characters' },
    { ok: /\p{L}/u.test(password), label: 'A letter' },
    { ok: /\d/.test(password), label: 'A number' },
  ];
}

// 0 (empty) to 4 (strong), only for the strength bar.
export function passwordStrength(password) {
  if (!password) return 0;
  let score = 0;
  if (password.length >= 8) score += 1;
  if (password.length >= 12) score += 1;
  if (/\p{L}/u.test(password) && /\d/.test(password)) score += 1;
  if (/[^\p{L}\d]/u.test(password) || (/\p{Lu}/u.test(password) && /\p{Ll}/u.test(password))) score += 1;
  return score;
}

export const STRENGTH_LABELS = ['', 'Weak', 'Fair', 'Good', 'Strong'];
