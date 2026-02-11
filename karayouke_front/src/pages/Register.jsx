import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../lib/auth.jsx';

const Register = () => {
	const [name, setName] = useState('');
	const [username, setUsername] = useState('');
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');
	const [confirmPassword, setConfirmPassword] = useState('');
	const [error, setError] = useState('');
	const [isLoading, setIsLoading] = useState(false);
	const { register } = useAuth();
	const navigate = useNavigate();

	const handleSubmit = async (e) => {
		e.preventDefault();
		setError('');

		if (password !== confirmPassword) {
			setError('Passwords do not match');
			return;
		}

		if (password.length < 6) {
			setError('Password must be at least 6 characters');
			return;
		}

		setIsLoading(true);

		try {
			await register(name, username, email, password);
			navigate('/', { replace: true });
		} catch (err) {
			setError(err.message);
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<div className="auth-page">
			<div className="auth-container">
				<div className="auth-card">
					<div className="auth-header">
						<div className="auth-logo">
							<span className="auth-logo-icon">♪</span>
						</div>
						<h1>Create account</h1>
						<p className="auth-subtitle">Join Karayouke and start hosting rooms</p>
					</div>

					<form onSubmit={handleSubmit} className="auth-form">
						{error && <div className="auth-error">{error}</div>}

						<div className="auth-field">
							<label htmlFor="name">Full Name</label>
							<input
								id="name"
								type="text"
								className="auth-input"
								placeholder="John Doe"
								value={name}
								onChange={(e) => setName(e.target.value)}
								required
								autoComplete="name"
							/>
						</div>

						<div className="auth-field">
							<label htmlFor="username">Username</label>
							<input
								id="username"
								type="text"
								className="auth-input"
								placeholder="johndoe"
								value={username}
								onChange={(e) => setUsername(e.target.value)}
								required
								autoComplete="username"
							/>
						</div>

						<div className="auth-field">
							<label htmlFor="email">Email</label>
							<input
								id="email"
								type="email"
								className="auth-input"
								placeholder="you@example.com"
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								required
								autoComplete="email"
							/>
						</div>

						<div className="auth-field">
							<label htmlFor="password">Password</label>
							<input
								id="password"
								type="password"
								className="auth-input"
								placeholder="••••••••"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								required
								autoComplete="new-password"
							/>
						</div>

						<div className="auth-field">
							<label htmlFor="confirmPassword">Confirm Password</label>
							<input
								id="confirmPassword"
								type="password"
								className="auth-input"
								placeholder="••••••••"
								value={confirmPassword}
								onChange={(e) => setConfirmPassword(e.target.value)}
								required
								autoComplete="new-password"
							/>
						</div>

						<button type="submit" className="auth-button" disabled={isLoading}>
							{isLoading ? (
								<span className="auth-spinner" />
							) : (
								'Create account'
							)}
						</button>
					</form>

					<div className="auth-footer">
						<p>
							Already have an account?{' '}
							<Link to="/login" className="auth-link">
								Sign in
							</Link>
						</p>
					</div>
				</div>

				<div className="auth-ambient" />
			</div>
		</div>
	);
};

export default Register;
