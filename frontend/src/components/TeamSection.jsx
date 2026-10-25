import { useState } from 'react';
import { ExternalLink, Mail } from 'lucide-react';
import { PROJECT, TEAM } from '../data/team';
import './TeamSection.css';

// Newer versions of lucide-react have no brand icons, so links are labelled pills.
const LINK_LABELS = {
  github: 'GitHub',
  linkedin: 'LinkedIn',
  facebook: 'Facebook',
  website: 'Website',
};

function initials(name) {
  return name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join('');
}

// Only real web addresses become links.
const isWebUrl = (value) => /^https?:\/\//i.test(value);

function Avatar({ member }) {
  const [broken, setBroken] = useState(false);

  if (!member.photo || broken) {
    return <div className="tm-avatar tm-avatar--initials" aria-hidden="true">{initials(member.name)}</div>;
  }
  return (
    <img
      className="tm-avatar"
      src={member.photo}
      alt={`Photo of ${member.name}`}
      loading="lazy"
      onError={() => setBroken(true)}
    />
  );
}

function MemberCard({ member }) {
  const links = Object.entries(LINK_LABELS)
    .filter(([key]) => isWebUrl(member.links?.[key] || ''))
    .map(([key, label]) => ({ key, label, href: member.links[key] }));

  const email = member.links?.email?.trim();

  return (
    <article className="tm-card">
      <Avatar member={member} />
      <div className="tm-body">
        <h3>{member.name}</h3>
        {member.role && <span className="tm-role">{member.role}</span>}
        <p className="tm-edu">{member.education}</p>
        <p className="tm-bio">{member.bio}</p>

      {(links.length > 0 || email) && (
        <div className="tm-links">
          {links.map(({ key, label, href }) => (
            <a key={key} href={href} target="_blank" rel="noopener noreferrer" aria-label={`${member.name} on ${label}`}>
              {label} <ExternalLink size={13} aria-hidden="true" />
            </a>
          ))}
          {email && (
            <a href={`mailto:${email}`} aria-label={`Email ${member.name}`}>
              <Mail size={14} aria-hidden="true" /> Email
            </a>
          )}
        </div>
      )}
      </div>
    </article>
  );
}

export default function TeamSection() {
  return (
    <section id="team" className="section tm" aria-labelledby="team-title">
      <div className="section-head tm-head">
        <h2 id="team-title">{PROJECT.heading}</h2>
        <p>{PROJECT.subheading}</p>
      </div>

      <div className="tm-grid">
        {TEAM.map((member) => (
          <MemberCard key={member.id} member={member} />
        ))}
      </div>
    </section>
  );
}
