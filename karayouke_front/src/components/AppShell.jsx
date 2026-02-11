import PropTypes from 'prop-types';

const AppShell = ({ title, subtitle, actions, children, className }) => (
	<div className={`app-shell${className ? ` ${className}` : ''}`}>
		<header className="app-header">
			<div>
				<p className="app-subtitle">{subtitle}</p>
				<h1 className="app-title">{title}</h1>
			</div>
			<div className="app-actions">{actions}</div>
		</header>
		<main className="app-content">{children}</main>
	</div>
);

AppShell.propTypes = {
	title: PropTypes.string.isRequired,
	subtitle: PropTypes.string,
	actions: PropTypes.node,
	children: PropTypes.node.isRequired,
	className: PropTypes.string,
};

AppShell.defaultProps = {
	subtitle: '',
	actions: null,
	className: '',
};

export default AppShell;
