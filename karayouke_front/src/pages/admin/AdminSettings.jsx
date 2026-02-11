import { useEffect, useState } from 'react';
import { fetchWithAuth } from '../../lib/auth.jsx';
import AdminLayout from './AdminLayout.jsx';

const API_BASE = (() => {
	const raw = import.meta.env.VITE_WS_HOST?.trim();
	if (raw && raw.length > 0) return raw.replace(/\/$/, '');
	if (typeof window !== 'undefined' && window.location) return window.location.origin.replace(/\/$/, '');
	return '';
})();

const AdminSettings = () => {
	const [configs, setConfigs] = useState([]);
	const [loading, setLoading] = useState(true);
	const [saving, setSaving] = useState({});
	const [error, setError] = useState(null);
	const [success, setSuccess] = useState(null);
	const [showAddForm, setShowAddForm] = useState(false);
	const [newConfig, setNewConfig] = useState({ key: '', value: '' });
	const [editingKey, setEditingKey] = useState(null);

	const configLabels = {
		room_max_duration: { label: 'Room Max Duration (minutes)', type: 'number', critical: true },
		room_creation_cost: { label: 'Room Creation Cost (credits)', type: 'number', critical: true },
		default_credits: { label: 'Default Credits for New Users', type: 'number', critical: true },
	};

	useEffect(() => {
		fetchConfigs();
	}, []);

	const fetchConfigs = async () => {
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/configs`);
			if (!response.ok) throw new Error('Failed to fetch configs');
			const data = await response.json();
			// Convert map to array format for easier handling
			const configArray = Object.entries(data || {}).map(([key, value]) => ({ key, value }));
			setConfigs(configArray);
		} catch (err) {
			setError(err.message);
		} finally {
			setLoading(false);
		}
	};

	const handleUpdate = async (key, value) => {
		setSaving((prev) => ({ ...prev, [key]: true }));
		setError(null);
		setSuccess(null);
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/configs/${key}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ value }),
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to update config');
			}
			setSuccess(`Successfully updated ${key}`);
			setEditingKey(null);
			await fetchConfigs();
		} catch (err) {
			setError(err.message);
		} finally {
			setSaving((prev) => ({ ...prev, [key]: false }));
		}
	};

	const handleCreate = async () => {
		if (!newConfig.key.trim() || !newConfig.value.trim()) {
			setError('Key and value are required');
			return;
		}
		setSaving((prev) => ({ ...prev, create: true }));
		setError(null);
		setSuccess(null);
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/configs`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(newConfig),
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to create config');
			}
			setSuccess(`Successfully created ${newConfig.key}`);
			setNewConfig({ key: '', value: '' });
			setShowAddForm(false);
			await fetchConfigs();
		} catch (err) {
			setError(err.message);
		} finally {
			setSaving((prev) => ({ ...prev, create: false }));
		}
	};

	const handleDelete = async (key) => {
		if (!confirm(`Are you sure you want to delete the config "${key}"?`)) {
			return;
		}
		setSaving((prev) => ({ ...prev, [key]: true }));
		setError(null);
		setSuccess(null);
		try {
			const response = await fetchWithAuth(`${API_BASE}/api/admin/configs/${key}`, {
				method: 'DELETE',
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to delete config');
			}
			setSuccess(`Successfully deleted ${key}`);
			await fetchConfigs();
		} catch (err) {
			setError(err.message);
		} finally {
			setSaving((prev) => ({ ...prev, [key]: false }));
		}
	};

	const getConfigValue = (key) => {
		const config = configs.find((c) => c.key === key);
		return config?.value || '';
	};

	return (
		<AdminLayout title="Settings">
			{loading ? (
				<div className="admin-loading-content">
					<span className="auth-spinner" />
				</div>
			) : (
				<div className="admin-settings">
					{error && (
						<div className="admin-error" style={{ marginBottom: '1rem' }}>
							{error}
							<button
								onClick={() => setError(null)}
								style={{ marginLeft: '1rem', cursor: 'pointer' }}
							>
								×
							</button>
						</div>
					)}
					{success && (
						<div
							className="admin-success"
							style={{
								marginBottom: '1rem',
								padding: '1rem',
								backgroundColor: '#d4edda',
								color: '#155724',
								borderRadius: '4px',
								border: '1px solid #c3e6cb',
							}}
						>
							{success}
							<button
								onClick={() => setSuccess(null)}
								style={{ marginLeft: '1rem', cursor: 'pointer' }}
							>
								×
							</button>
						</div>
					)}

					<div className="admin-card">
						<div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
							<div>
								<h2 className="admin-card-title">System Configuration</h2>
								<p className="admin-card-subtitle">
									Configure global settings for the karaoke system
								</p>
							</div>
							<button
								className="admin-btn admin-btn-primary"
								onClick={() => setShowAddForm(!showAddForm)}
							>
								{showAddForm ? 'Cancel' : '+ Add Config'}
							</button>
						</div>

						{showAddForm && (
							<div
								className="admin-add-form"
								style={{
									padding: '1rem',
									backgroundColor: '#f8f9fa',
									borderRadius: '4px',
									marginBottom: '1rem',
								}}
							>
								<h3 style={{ marginBottom: '1rem' }}>Add New Configuration</h3>
								<div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
									<input
										type="text"
										className="admin-input"
										placeholder="Config key (e.g., custom_setting)"
										value={newConfig.key}
										onChange={(e) =>
											setNewConfig({ ...newConfig, key: e.target.value })
										}
										style={{ flex: 1 }}
									/>
									<input
										type="text"
										className="admin-input"
										placeholder="Config value"
										value={newConfig.value}
										onChange={(e) =>
											setNewConfig({ ...newConfig, value: e.target.value })
										}
										style={{ flex: 1 }}
									/>
								</div>
								<button
									className="admin-btn admin-btn-primary"
									onClick={handleCreate}
									disabled={saving.create}
								>
									{saving.create ? 'Creating...' : 'Create Config'}
								</button>
							</div>
						)}

						<div className="admin-settings-list">
							{configs.map(({ key, value }) => {
								const config = configLabels[key] || { label: key, type: 'text', critical: false };
								const isEditing = editingKey === key;
								const currentValue = getConfigValue(key);

								return (
									<div key={key} className="admin-setting-item">
										<div style={{ flex: 1 }}>
											<label className="admin-setting-label">
												{config.label}
												{config.critical && (
													<span
														style={{
															marginLeft: '0.5rem',
															fontSize: '0.75rem',
															color: '#dc3545',
															fontWeight: 'bold',
														}}
													>
														(Critical)
													</span>
												)}
											</label>
											<div className="admin-setting-control">
												<input
													type={config.type}
													className="admin-input"
													value={currentValue}
													onChange={(e) => {
														const newConfigs = configs.map((c) =>
															c.key === key ? { ...c, value: e.target.value } : c
														);
														setConfigs(newConfigs);
														setEditingKey(key);
													}}
													style={{ flex: 1 }}
												/>
												<div style={{ display: 'flex', gap: '0.5rem' }}>
													{isEditing && (
														<button
															className="admin-btn admin-btn-primary"
															onClick={() => handleUpdate(key, currentValue)}
															disabled={saving[key]}
														>
															{saving[key] ? 'Saving...' : 'Save'}
														</button>
													)}
													{!config.critical && (
														<button
															className="admin-btn"
															style={{
																backgroundColor: '#dc3545',
																color: 'white',
															}}
															onClick={() => handleDelete(key)}
															disabled={saving[key]}
														>
															{saving[key] ? 'Deleting...' : 'Delete'}
														</button>
													)}
												</div>
											</div>
										</div>
									</div>
								);
							})}
						</div>
					</div>

					<div className="admin-card">
						<h2 className="admin-card-title">Configuration Guide</h2>
						<div className="admin-info-list">
							<div className="admin-info-item">
								<strong>Room Max Duration:</strong> How long a room stays active after
								creation (in minutes). After this time, the room expires and users are
								kicked.
							</div>
							<div className="admin-info-item">
								<strong>Room Creation Cost:</strong> Number of credits required to create
								a new room. Set to 0 for free room creation.
							</div>
							<div className="admin-info-item">
								<strong>Default Credits:</strong> Credits given to new users upon
								registration. Used for initial room creation.
							</div>
							<div className="admin-info-item">
								<strong>Critical Configs:</strong> System-critical configurations cannot be
								deleted to prevent breaking the application. You can only update their values.
							</div>
						</div>
					</div>
				</div>
			)}
		</AdminLayout>
	);
};

export default AdminSettings;
