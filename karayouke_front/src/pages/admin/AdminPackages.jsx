import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../../lib/auth.jsx';
import { formatIDR } from '../../lib/currency.jsx';
import AdminLayout from './AdminLayout.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const AdminPackages = () => {
	const [packages, setPackages] = useState([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState(null);
	const [showModal, setShowModal] = useState(false);
	const [editingPackage, setEditingPackage] = useState(null);
	const [formData, setFormData] = useState({
		package_name: '',
		package_detail: '',
		credit_amount: '',
		price: '',
		visibility: true,
	});
	const [submitting, setSubmitting] = useState(false);

	useEffect(() => {
		fetchPackages();
	}, []);

	const fetchPackages = async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/packages`);
			if (!response.ok) throw new Error('Failed to fetch packages');
			const data = await response.json();
			setPackages(Array.isArray(data) ? data : []);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	};

	const openCreateModal = () => {
		setEditingPackage(null);
		setFormData({ package_name: '', package_detail: '', credit_amount: '', price: '', visibility: true });
		setShowModal(true);
	};

	const openEditModal = (pkg) => {
		setEditingPackage(pkg);
		// Handle package_detail as bytes/string from Go
		let detail = '';
		if (pkg.package_detail) {
			if (typeof pkg.package_detail === 'string') {
				detail = pkg.package_detail;
			} else if (Array.isArray(pkg.package_detail)) {
				// It's a byte array from Go, decode it
				detail = String.fromCharCode(...pkg.package_detail);
			}
		}
		setFormData({
			package_name: pkg.package_name || '',
			package_detail: detail,
			credit_amount: (pkg.credit_amount || 0).toString(),
			price: (pkg.price || 0).toString(),
			visibility: pkg.visibility !== false,
		});
		setShowModal(true);
	};

	const handleSubmit = async (e) => {
		e.preventDefault();
		setSubmitting(true);
		try {
			const payload = {
				package_name: formData.package_name,
				package_detail: formData.package_detail,
				credit_amount: parseInt(formData.credit_amount, 10),
				price: parseInt(formData.price, 10) || 0,
				visibility: formData.visibility,
			};

			const url = editingPackage
				? `${API_BASE}/api/admin/packages/${editingPackage.id}`
				: `${API_BASE}/api/admin/packages`;

			const response = await fetchWithAuth(url, {
				method: editingPackage ? 'PUT' : 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(payload),
			});

			if (!response.ok) throw new Error('Failed to save package');
			await fetchPackages();
			setShowModal(false);
		} catch (err) {
			setError(err.message);
		} finally {
			setSubmitting(false);
		}
	};

	const handleDelete = async (id) => {
		if (!confirm('Are you sure you want to delete this package?')) return;
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/packages/${id}`, {
				method: 'DELETE',
			});
			if (!response.ok) throw new Error('Failed to delete package');
			await fetchPackages();
		} catch (err) {
			setError(err.message);
		}
	};

	// Helper to decode package_detail from byte array
	const getPackageDetail = (pkg) => {
		if (!pkg.package_detail) return '';
		if (typeof pkg.package_detail === 'string') return pkg.package_detail;
		if (Array.isArray(pkg.package_detail)) {
			return String.fromCharCode(...pkg.package_detail);
		}
		return '';
	};

	return (
		<AdminLayout title="Packages">
			{loading ? (
				<div className="admin-loading-content">
					<span className="auth-spinner" />
				</div>
			) : (
				<div className="admin-packages">
					<div className="admin-toolbar">
						<button className="admin-btn admin-btn-primary" onClick={openCreateModal}>
							+ Create Package
						</button>
					</div>

					{error && <div className="admin-error">{error}</div>}

					<div className="admin-table-container">
						<table className="admin-table">
							<thead>
								<tr>
									<th>Name</th>
									<th>Credits</th>
									<th>Price (IDR)</th>
									<th>Visible</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{packages.length === 0 ? (
									<tr>
										<td colSpan="5" className="admin-table-empty">
											No packages found. Create one to get started.
										</td>
									</tr>
								) : (
									packages.map((pkg) => (
										<tr key={pkg.id}>
											<td>
												<div className="admin-cell-primary">{pkg.package_name}</div>
												{getPackageDetail(pkg) && (
													<div className="admin-cell-secondary">{getPackageDetail(pkg)}</div>
												)}
											</td>
											<td>{pkg.credit_amount}</td>
											<td>{formatIDR(pkg.price)}</td>
											<td>
												<span
													className={`admin-badge ${pkg.visibility ? 'success' : 'muted'}`}
												>
													{pkg.visibility ? 'Yes' : 'No'}
												</span>
											</td>
											<td>
												<div className="admin-actions">
													<button
														className="admin-btn admin-btn-sm"
														onClick={() => openEditModal(pkg)}
													>
														Edit
													</button>
													<button
														className="admin-btn admin-btn-sm admin-btn-danger"
														onClick={() => handleDelete(pkg.id)}
													>
														Delete
													</button>
												</div>
											</td>
										</tr>
									))
								)}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{showModal && (
				<div className="admin-modal-overlay" onClick={() => setShowModal(false)}>
					<div className="admin-modal" onClick={(e) => e.stopPropagation()}>
						<div className="admin-modal-header">
							<h3>{editingPackage ? 'Edit Package' : 'Create Package'}</h3>
							<button className="admin-modal-close" onClick={() => setShowModal(false)}>
								×
							</button>
						</div>
						<form onSubmit={handleSubmit} className="admin-modal-body">
							<div className="admin-form-group">
								<label>Name</label>
								<input
									type="text"
									className="admin-input"
									value={formData.package_name}
									onChange={(e) => setFormData({ ...formData, package_name: e.target.value })}
									required
								/>
							</div>
							<div className="admin-form-group">
								<label>Description</label>
								<textarea
									className="admin-input admin-textarea"
									value={formData.package_detail}
									onChange={(e) =>
										setFormData({ ...formData, package_detail: e.target.value })
									}
								/>
							</div>
							<div className="admin-form-row">
								<div className="admin-form-group">
									<label>Credits</label>
									<input
										type="number"
										className="admin-input"
										value={formData.credit_amount}
										onChange={(e) => setFormData({ ...formData, credit_amount: e.target.value })}
										required
										min="1"
									/>
								</div>
								<div className="admin-form-group">
									<label>Price (IDR)</label>
									<input
										type="number"
										className="admin-input"
										value={formData.price}
										onChange={(e) => setFormData({ ...formData, price: e.target.value })}
										required
										min="0"
										step="1000"
										placeholder="e.g., 50000"
									/>
								</div>
							</div>
							<div className="admin-form-group">
								<label className="admin-checkbox-label">
									<input
										type="checkbox"
										checked={formData.visibility}
										onChange={(e) =>
											setFormData({ ...formData, visibility: e.target.checked })
										}
									/>
									<span>Visible to users</span>
								</label>
							</div>
							<div className="admin-modal-footer">
								<button
									type="button"
									className="admin-btn"
									onClick={() => setShowModal(false)}
								>
									Cancel
								</button>
								<button
									type="submit"
									className="admin-btn admin-btn-primary"
									disabled={submitting}
								>
									{submitting ? 'Saving...' : editingPackage ? 'Update' : 'Create'}
								</button>
							</div>
						</form>
					</div>
				</div>
			)}
		</AdminLayout>
	);
};

export default AdminPackages;
