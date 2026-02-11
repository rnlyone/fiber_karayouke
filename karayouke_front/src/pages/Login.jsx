import { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../lib/auth.jsx';

const Login = () => {
	const [email, setEmail] = useState('');
	const [password, setPassword] = useState('');
	const [error, setError] = useState('');
	const [isLoading, setIsLoading] = useState(false);
	const { login } = useAuth();
	const navigate = useNavigate();
	const location = useLocation();

	const from = location.state?.from || '/';

	const handleSubmit = async (e) => {
		e.preventDefault();
		setError('');
		setIsLoading(true);

		try {
			await login(email, password);
			navigate(from, { replace: true });
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
						<h1>Welcome back</h1>
						<p className="auth-subtitle">Sign in to continue to Karayouke</p>
					</div>

					<form onSubmit={handleSubmit} className="auth-form">
						{error && <div className="auth-error">{error}</div>}

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
								autoComplete="current-password"
							/>
						</div>

						<button type="submit" className="auth-button" disabled={isLoading}>
							{isLoading ? (
								<span className="auth-spinner" />
							) : (
								'Sign in'
							)}
						</button>
					</form>

					<div className="auth-footer">
						<p>
							Don't have an account?{' '}
							<Link to="/register" className="auth-link">
								Create one
							</Link>
						</p>
					</div>
				</div>

				<div className="auth-ambient" />
			</div>
		</div>
	);
};

export default Login;
