import { Link } from 'react-router-dom';

const faqs = [
	{
		question: 'What is Karayouke?',
		answer: 'Karayouke is a collaborative karaoke platform that lets you create rooms, invite friends, and queue songs together in real-time. One device plays the video while everyone uses their phones to search and add songs to the queue.',
	},
	{
		question: 'How do I create a karaoke room?',
		answer: 'Sign in to your account, go to the Dashboard, enter a room name and click "Create". You will need credits to create a room. Once created, share the room code or QR code with your friends.',
	},
	{
		question: 'How do guests join a room?',
		answer: 'Guests can join by scanning the QR code displayed on the player screen, or by visiting the "Join a Room" page and entering the room code. No account is required to join as a guest.',
	},
	{
		question: 'What are credits and how do I get them?',
		answer: 'Credits are used to create karaoke rooms. Each room creation costs credits. You can purchase credit packages from the Packages page using bank transfer, virtual account, QRIS, e-wallet, and other Indonesian payment methods via Flip.',
	},
	{
		question: 'How long does a room last?',
		answer: 'Room duration depends on your package. Once a room expires, all connected users will be notified and the session will end. You can see the remaining time on your Dashboard.',
	},
	{
		question: 'Can multiple people add songs at the same time?',
		answer: 'Yes! That is the core feature of Karayouke. Multiple guests can search and queue songs simultaneously from their own devices. The queue updates in real-time for everyone.',
	},
	{
		question: 'What is the "Play Next" button?',
		answer: 'The "Play Next" button inserts a song right after the currently playing song, rather than at the end of the queue. The "+" button adds the song to the end of the queue.',
	},
	{
		question: 'Can I reorder the song queue?',
		answer: 'Yes, the room master and controllers can drag and drop songs in the queue to reorder them. On mobile devices, you can use the up/down arrow buttons.',
	},
	{
		question: 'What devices are supported?',
		answer: 'Karayouke works on any modern web browser. The player screen works great on smart TVs (Samsung Tizen, LG webOS), computers, and tablets. The controller works on any smartphone or computer.',
	},
	{
		question: 'Is my payment information secure?',
		answer: 'Yes. All payments are processed through Flip, a trusted Indonesian payment gateway. We never store your payment details on our servers.',
	},
	{
		question: 'How do I contact support?',
		answer: 'You can reach us by email at ask@karayouke.com. We will respond as soon as possible.',
	},
];

const FAQ = () => {
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
				<h1>Frequently Asked Questions</h1>
				<div className="faq-list">
					{faqs.map((faq, index) => (
						<details key={index} className="faq-item">
							<summary className="faq-question">{faq.question}</summary>
							<p className="faq-answer">{faq.answer}</p>
						</details>
					))}
				</div>
				<div className="legal-contact">
					<h2>Still have questions?</h2>
					<p>Contact us at <a href="mailto:ask@karayouke.com">ask@karayouke.com</a></p>
					<p className="legal-address">Tamangapa Raya No. 43, Bangkala, Manggala, Kota Makassar, Sulawesi Selatan, Indonesia 90235</p>
				</div>
			</main>
		</div>
	);
};

export default FAQ;
