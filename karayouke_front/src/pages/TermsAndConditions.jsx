import { Link } from 'react-router-dom';

const TermsAndConditions = () => {
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
				<h1>Terms and Conditions</h1>
				<p className="legal-updated">Last updated: February 11, 2026</p>

				<section className="legal-section">
					<h2>1. Acceptance of Terms</h2>
					<p>
						By accessing and using Karayouke (karayouke.com), you agree to be bound by these Terms and Conditions. 
						If you do not agree with any part of these terms, you may not use our service.
					</p>
				</section>

				<section className="legal-section">
					<h2>2. Description of Service</h2>
					<p>
						Karayouke is a web-based collaborative karaoke platform that allows users to create rooms, 
						invite guests, and manage song queues in real-time. The service streams karaoke videos from 
						YouTube and provides tools for collaborative playlist management.
					</p>
				</section>

				<section className="legal-section">
					<h2>3. User Accounts</h2>
					<ul>
						<li>You must provide accurate and complete information when creating an account.</li>
						<li>You are responsible for maintaining the confidentiality of your account credentials.</li>
						<li>You are responsible for all activities that occur under your account.</li>
						<li>You must notify us immediately of any unauthorized use of your account.</li>
						<li>You must be at least 13 years of age to create an account.</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>4. Credits and Payments</h2>
					<ul>
						<li>Credits are required to create karaoke rooms.</li>
						<li>All payments are processed securely through Stripe.</li>
						<li>Prices are displayed in your detected currency and charged accordingly.</li>
						<li>Credits are non-transferable between accounts.</li>
						<li>We reserve the right to change credit pricing at any time with prior notice.</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>5. Acceptable Use</h2>
					<p>When using Karayouke, you agree not to:</p>
					<ul>
						<li>Use the service for any illegal or unauthorized purpose.</li>
						<li>Attempt to gain unauthorized access to any part of the service.</li>
						<li>Interfere with or disrupt the service or servers.</li>
						<li>Use automated scripts to access or interact with the service.</li>
						<li>Share content that is offensive, harmful, or violates third-party rights.</li>
						<li>Attempt to reverse engineer or decompile the service.</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>6. Intellectual Property</h2>
					<p>
						The Karayouke name, logo, and all related content, features, and functionality are owned by 
						Karayouke and are protected by intellectual property laws. Video content is streamed via YouTube 
						and is subject to YouTube's Terms of Service.
					</p>
				</section>

				<section className="legal-section">
					<h2>7. Third-Party Services</h2>
					<p>
						Karayouke integrates with third-party services including YouTube (for video content) and 
						Stripe (for payment processing). Your use of these services is subject to their respective 
						terms of service and privacy policies.
					</p>
				</section>

				<section className="legal-section">
					<h2>8. Room Usage</h2>
					<ul>
						<li>Rooms are created using credits and have a limited duration based on the package.</li>
						<li>Room creators are responsible for the content queued in their rooms.</li>
						<li>We reserve the right to terminate any room that violates these terms.</li>
						<li>Expired rooms and their data may be deleted after 30 days.</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>9. Disclaimer of Warranties</h2>
					<p>
						The service is provided "as is" and "as available" without warranties of any kind, 
						either express or implied. We do not guarantee that the service will be uninterrupted, 
						error-free, or free from harmful components.
					</p>
				</section>

				<section className="legal-section">
					<h2>10. Limitation of Liability</h2>
					<p>
						To the fullest extent permitted by law, Karayouke shall not be liable for any indirect, 
						incidental, special, consequential, or punitive damages resulting from your use of or 
						inability to use the service, including but not limited to loss of data, loss of profits, 
						or service interruptions.
					</p>
				</section>

				<section className="legal-section">
					<h2>11. Termination</h2>
					<p>
						We may suspend or terminate your access to the service at any time, with or without cause, 
						and with or without notice. Upon termination, your right to use the service will immediately cease.
					</p>
				</section>

				<section className="legal-section">
					<h2>12. Governing Law</h2>
					<p>
						These terms shall be governed by and construed in accordance with the laws of the Republic of Indonesia, 
						without regard to conflict of law principles.
					</p>
				</section>

				<section className="legal-section">
					<h2>13. Changes to Terms</h2>
					<p>
						We reserve the right to modify these terms at any time. Changes will be posted on this page 
						with an updated revision date. Continued use of the service after changes constitutes acceptance of the new terms.
					</p>
				</section>

				<div className="legal-contact">
					<h2>Contact Us</h2>
					<p>If you have any questions about these terms, please contact us:</p>
					<p><strong>Email:</strong> <a href="mailto:ask@karayouke.com">ask@karayouke.com</a></p>
					<p className="legal-address">Tamangapa Raya No. 43, Bangkala, Manggala, Kota Makassar, Sulawesi Selatan, Indonesia 90235</p>
				</div>
			</main>
		</div>
	);
};

export default TermsAndConditions;
