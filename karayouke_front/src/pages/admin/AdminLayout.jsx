import { useEffect, useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../lib/auth.jsx';

const ADMIN_EMAIL = 'me@ruzman.my.id';

const AdminLayout = ({ children, title }) => {
	const { user, isAuthenticated, isLoading, logout } = useAuth();
	const navigate = useNavigate();
	const location = useLocation();
	const [sidebarOpen, setSidebarOpen] = useState(false);
	const [userMenuOpen, setUserMenuOpen] = useState(false);

	const isAdmin = user?.email === ADMIN_EMAIL;

	useEffect(() => {
		if (!isLoading && (!isAuthenticated || !isAdmin)) {
			navigate('/');
		}
	}, [isLoading, isAuthenticated, isAdmin, navigate]);

	const handleLogout = async () => {
		await logout();
		navigate('/login');
	};

	if (isLoading) {
		return (
			<div className="admin-page">
				<div className="admin-loading">
					<span className="auth-spinner" />
				</div>
			</div>
		);
	}

	if (!isAuthenticated || !isAdmin) {
		return null;
	}

	const navItems = [
		{ path: '/admin', label: 'Dashboard', icon: '📊' },
		{ path: '/admin/settings', label: 'Settings', icon: '⚙️' },
		{ path: '/admin/packages', label: 'Packages', icon: '📦' },
		{ path: '/admin/users', label: 'Users', icon: '👥' },
		{ path: '/admin/transactions', label: 'Transactions', icon: '💳' },
		{ path: '/admin/rooms', label: 'Rooms', icon: '🎤' },
	];

	return (
		<div className="admin-page">
			<aside className={`admin-sidebar ${sidebarOpen ? 'open' : ''}`}>
				<div className="admin-sidebar-header">
					<Link to="/" className="admin-logo">
						<span className="admin-logo-icon">♪</span>
						<span className="admin-logo-text brand-logo-text">Karayouke</span>
					</Link>
					<button className="admin-sidebar-close" onClick={() => setSidebarOpen(false)}>
						×
					</button>
				</div>
				<nav className="admin-nav">
					{navItems.map((item) => (
						<Link
							key={item.path}
							to={item.path}
							className={`admin-nav-item ${location.pathname === item.path ? 'active' : ''}`}
							onClick={() => setSidebarOpen(false)}
						>
							<span className="admin-nav-icon">{item.icon}</span>
							<span className="admin-nav-label">{item.label}</span>
						</Link>
					))}
				</nav>
				<div className="admin-sidebar-footer">
					<Link to="/" className="admin-nav-item">
						<span className="admin-nav-icon">🏠</span>
						<span className="admin-nav-label">Back to App</span>
					</Link>
				</div>
			</aside>

			<div className="admin-main">
				<header className="admin-header">
					<button className="admin-menu-btn" onClick={() => setSidebarOpen(true)}>
						☰
					</button>
					<h1 className="admin-header-title">{title}</h1>
					<div className="admin-header-user">
						<button
							className="admin-user-menu-btn"
							onClick={() => setUserMenuOpen((prev) => !prev)}
							aria-label="Toggle user menu"
							aria-expanded={userMenuOpen}
						>
							⋮
						</button>
						<div className={`admin-header-user-items ${userMenuOpen ? 'open' : ''}`}>
							<span className="admin-user-name">{user?.name}</span>
							<button className="admin-logout-btn" onClick={handleLogout}>
								Sign out
							</button>
						</div>
					</div>
				</header>
				<main className="admin-content">{children}</main>
			</div>
		</div>
	);
};

export default AdminLayout;
