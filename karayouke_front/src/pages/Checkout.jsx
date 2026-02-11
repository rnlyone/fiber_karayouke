import { useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth, getAuthToken } from '../lib/auth.jsx';

/**
 * Checkout is now just a redirect helper.
 * The actual purchase flow is handled on the Packages page
 * which calls iPaymu create-payment and redirects to iPaymu's hosted page.
 * This page handles the case where someone directly navigates to /checkout/:packageId.
 */
const Checkout = () => {
	const { packageId } = useParams();
	const navigate = useNavigate();
	const { isAuthenticated, isLoading: authLoading } = useAuth();

	useEffect(() => {
		const hasToken = !!getAuthToken();
		if (!authLoading && !isAuthenticated && !hasToken) {
			navigate('/login', { state: { from: `/checkout/${packageId}` } });
			return;
		}
		// Redirect to packages page since purchases are initiated there
		if (isAuthenticated || hasToken) {
			navigate('/packages', { replace: true });
		}
	}, [packageId, isAuthenticated, authLoading, navigate]);

	return (
		<div className="checkout-page">
			<div className="checkout-loading">
				<span className="auth-spinner" />
				<p>Redirecting to packages...</p>
				<Link to="/packages">Go to Packages</Link>
			</div>
		</div>
	);
};

export default Checkout;
