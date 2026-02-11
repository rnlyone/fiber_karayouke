import { useState, useEffect, useCallback, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { QRCodeCanvas } from 'qrcode.react';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const TVPage = () => {
	const navigate = useNavigate();
	const [token, setToken] = useState(null);
	const [shortCode, setShortCode] = useState(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [countdown, setCountdown] = useState(60);
	const pollIntervalRef = useRef(null);
	const refreshIntervalRef = useRef(null);

	// Generate QR URL for the token
	const qrUrl = token ? `${window.location.origin}/tv/connect/${token}` : '';

	// Generate new token
	const generateToken = useCallback(async () => {
		setLoading(true);
		setError(null);
		try {
			const response = await fetch(`${API_BASE}/api/tv/token`, {
				method: 'POST',
			});
			if (!response.ok) {
				throw new Error('Failed to generate token');
			}
			const data = await response.json();
			setToken(data.token);
			setShortCode(data.short_code);
			setCountdown(60);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	}, []);

	// Poll for connection status
	const checkStatus = useCallback(async () => {
		if (!token) return;
		try {
			const response = await fetch(`${API_BASE}/api/tv/status/${token}`);
			const data = await response.json();
			
			if (data.expired) {
				// Token expired, generate new one
				generateToken();
				return;
			}

			if (data.connected && data.room_key) {
				// Connected! Navigate to player
				navigate(`/rooms/${data.room_key}/player`, { replace: true });
			}
		} catch {
			// Ignore errors, will retry
		}
	}, [token, navigate, generateToken]);

	// Initial token generation
	useEffect(() => {
		generateToken();
	}, [generateToken]);

	// Poll for status every 2 seconds
	useEffect(() => {
		if (token) {
			pollIntervalRef.current = setInterval(checkStatus, 2000);
		}
		return () => {
			if (pollIntervalRef.current) {
				clearInterval(pollIntervalRef.current);
			}
		};
	}, [token, checkStatus]);

	// Auto-refresh token every 60 seconds
	useEffect(() => {
		refreshIntervalRef.current = setInterval(() => {
			generateToken();
		}, 60000);
		return () => {
			if (refreshIntervalRef.current) {
				clearInterval(refreshIntervalRef.current);
			}
		};
	}, [generateToken]);

	// Countdown timer
	useEffect(() => {
		const timer = setInterval(() => {
			setCountdown((prev) => {
				if (prev <= 1) return 60;
				return prev - 1;
			});
		}, 1000);
		return () => clearInterval(timer);
	}, []);

	if (loading && !token) {
		return (
			<div className="tv-page">
				<div className="tv-loading">
					<span className="auth-spinner" />
					<p>Preparing TV connection...</p>
				</div>
			</div>
		);
	}

	if (error) {
		return (
			<div className="tv-page">
				<div className="tv-error">
					<h2>Connection Error</h2>
					<p>{error}</p>
					<button onClick={generateToken} className="tv-retry-btn">
						Try Again
					</button>
				</div>
			</div>
		);
	}

	return (
		<div className="tv-page">
			<div className="tv-content">
				<div className="tv-connect-section">
					<div className="tv-card-header">
						<div className="tv-card-logo">
							<span className="tv-card-logo-icon" aria-hidden="true">🎤</span>
							<div className="tv-card-logo-text">
								<span className="tv-card-logo-title brand-logo-text">Karayouke</span>
								<span className="tv-card-logo-subtitle">TV Player</span>
							</div>
						</div>
						<h2>Follow these steps on your computer or phone</h2>
					</div>

					<div className="tv-connect-methods">
						{/* QR Code - Left Side */}
						<div className="tv-qr-section">
							<div className="tv-qr-wrapper">
								<QRCodeCanvas
									value={qrUrl}
									size={280}
									fgColor="#000"
									bgColor="#fff"
									includeMargin
									level="H"
								/>
							</div>
						</div>

						{/* Steps - Right Side */}
						<div className="tv-steps">
							<div className="tv-step">
								<div className="tv-step-number">STEP 1</div>
								<div className="tv-step-content">
									<h3>Login & Open Your Room Page</h3>
									<p className="tv-step-url">{window.location.origin}</p>
								</div>
							</div>

							<div className="tv-step">
								<div className="tv-step-number">STEP 2</div>
								<div className="tv-step-content">
									<h3>Scan This QR <b>or</b> Enter this code</h3>
									<div className="tv-short-code">
										{shortCode?.split('').map((char, i) => (
											<span key={i} className="tv-code-char">{char}</span>
										))}
									</div>
								</div>
							</div>

							<div className="tv-step">
								<div className="tv-step-number">STEP 3</div>
								<div className="tv-step-content">
									<h3>Sign in and start singing!</h3>
								</div>
							</div>
                            <div className="tv-status-row">
                                <div className="tv-status-spacer" aria-hidden="true" />
                                <div className="tv-status-box">
                                    <div className="tv-countdown">
                                        <span className="tv-countdown-icon">.</span>
                                        <span>&nbsp;Code refreshes in {countdown}s</span>
                                    </div>
                                    <div className="tv-status-wait">
                                        <span>Waiting for connection...</span>
                                        <div className="tv-pulse" />
                                    </div>
                                </div>
                            </div>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
};

export default TVPage;
