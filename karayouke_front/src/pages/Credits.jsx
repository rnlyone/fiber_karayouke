import { Link } from 'react-router-dom';

const Credits = () => {
	const version = typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : '0.0.0';

	return (
		<div className="legal-page">
			<header className="legal-header">
				<div className="legal-brand">
					<Link to="/" className="legal-back">←</Link>
					<span className="legal-logo">♪</span>
					<span className="legal-title">Karayouke</span>
				</div>
			</header>
			<main className="legal-main">
				<div className="credits-hero">
					<span className="credits-icon">♪</span>
					<h1>Karayouke</h1>
					<p className="credits-tagline">Collaborative Karaoke Platform</p>
					<span className="credits-version">v{version}</span>
				</div>

				<section className="credits-section">
					<h2>Created by</h2>
					<div className="credits-card">
						<div className="credits-avatar">R</div>
						<div className="credits-info">
							<p className="credits-name">Ruzman</p>
							<a href="mailto:me@ruzman.my.id" className="credits-email">me@ruzman.my.id</a>
						</div>
					</div>
				</section>

				<section className="credits-section">
					<h2>Technology</h2>
					<div className="credits-tech-grid">
						<div className="credits-tech">
							<span className="credits-tech-icon">⚡</span>
							<span>Go &amp; Fiber</span>
						</div>
						<div className="credits-tech">
							<span className="credits-tech-icon">⚛</span>
							<span>React</span>
						</div>
						<div className="credits-tech">
							<span className="credits-tech-icon">🔌</span>
							<span>WebSocket</span>
						</div>
						<div className="credits-tech">
							<span className="credits-tech-icon">▶</span>
							<span>YouTube API</span>
						</div>
						<div className="credits-tech">
							<span className="credits-tech-icon">💳</span>
							<span>Stripe</span>
						</div>
						<div className="credits-tech">
							<span className="credits-tech-icon">🐘</span>
							<span>PostgreSQL</span>
						</div>
					</div>
				</section>

				<section className="credits-section">
					<h2>Contact</h2>
					<div className="credits-contact-list">
						<p>
							<strong>Business inquiries:</strong>{' '}
							<a href="mailto:ask@karayouke.com">ask@karayouke.com</a>
						</p>
						<p>
							<strong>Personal:</strong>{' '}
							<a href="mailto:me@ruzman.my.id">me@ruzman.my.id</a>
						</p>
					</div>
				</section>

				<div className="credits-footer">
					<p>&copy; {new Date().getFullYear()} Karayouke. All rights reserved.</p>
					<div className="credits-links">
						<Link to="/faq">FAQ</Link>
						<Link to="/terms">Terms</Link>
						<Link to="/refund-policy">Refund Policy</Link>
					</div>
				</div>
			</main>
		</div>
	);
};

export default Credits;
