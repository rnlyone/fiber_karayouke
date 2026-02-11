import { Link } from 'react-router-dom';

const RefundPolicy = () => {
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
				<h1>Refund Policy</h1>
				<p className="legal-updated">Last updated: February 11, 2026</p>

				<section className="legal-section">
					<h2>1. Overview</h2>
					<p>
						This Refund Policy applies to all credit purchases made on Karayouke (karayouke.com). 
						We want you to be satisfied with your purchase. Please read this policy carefully before making a purchase.
					</p>
				</section>

				<section className="legal-section">
					<h2>2. Digital Products</h2>
					<p>
						All purchases on Karayouke are for digital credits used to create karaoke rooms. 
						As digital products, credits are delivered instantly upon successful payment.
					</p>
				</section>

				<section className="legal-section">
					<h2>3. Refund Eligibility</h2>
					<p>You may be eligible for a refund in the following cases:</p>
					<ul>
						<li><strong>Technical failure:</strong> If you were charged but credits were not added to your account due to a technical error.</li>
						<li><strong>Duplicate charge:</strong> If you were charged multiple times for the same purchase.</li>
						<li><strong>Unauthorized transaction:</strong> If a purchase was made without your authorization.</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>4. Non-Refundable Cases</h2>
					<p>Refunds will not be granted in the following cases:</p>
					<ul>
						<li>Credits that have already been used to create rooms.</li>
						<li>Dissatisfaction with the service after credits have been consumed.</li>
						<li>Failure to use credits before they expire (if applicable).</li>
						<li>Issues caused by your own internet connection or device.</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>5. How to Request a Refund</h2>
					<p>
						To request a refund, please contact us at <a href="mailto:ask@karayouke.com">ask@karayouke.com</a> with the following information:
					</p>
					<ul>
						<li>Your account email address</li>
						<li>Transaction ID or payment receipt</li>
						<li>Reason for the refund request</li>
						<li>Date of purchase</li>
					</ul>
				</section>

				<section className="legal-section">
					<h2>6. Refund Processing</h2>
					<p>
						Approved refunds will be processed within 7-14 business days. 
						Refunds will be returned to the original payment method used for the purchase. 
						Processing times may vary depending on your bank or card issuer.
					</p>
				</section>

				<section className="legal-section">
					<h2>7. Partial Refunds</h2>
					<p>
						In some cases, we may offer a partial refund or credit adjustment instead of a full refund, 
						depending on the circumstances and how many credits have already been used.
					</p>
				</section>

				<section className="legal-section">
					<h2>8. Changes to This Policy</h2>
					<p>
						We reserve the right to modify this refund policy at any time. 
						Changes will be posted on this page with an updated revision date.
					</p>
				</section>

				<div className="legal-contact">
					<h2>Contact Us</h2>
					<p>If you have any questions about this refund policy, please contact us:</p>
					<p><strong>Email:</strong> <a href="mailto:ask@karayouke.com">ask@karayouke.com</a></p>
					<p className="legal-address">Tamangapa Raya No. 43, Bangkala, Manggala, Kota Makassar, Sulawesi Selatan, Indonesia 90235</p>
				</div>
			</main>
		</div>
	);
};

export default RefundPolicy;
