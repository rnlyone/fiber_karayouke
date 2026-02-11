import PropTypes from 'prop-types';

const RoomControllerPanel = ({ queue, nowPlaying, onMove, onRemove, onSkip }) => {
	return (
		<div className="card">
			<div className="card-header">
				<h3>Queue control</h3>
				<p className="text-muted">Reorder, remove, or skip songs in real time.</p>
			</div>
			{nowPlaying ? (
				<div className="media">
					<img src={nowPlaying.thumbnail || 'https://placehold.co/112x112?text=Now'} alt={nowPlaying.title} />
					<div>
						<p className="list-title">{nowPlaying.title}</p>
						<p className="text-muted">Now playing</p>
					</div>
				</div>
			) : (
				<p className="text-muted">Nothing playing yet.</p>
			)}
			<div className="divider" />
			{queue.length === 0 ? (
				<p className="text-muted">Queue is empty. Ask guests to add songs.</p>
			) : (
				<div className="queue-list">
					{queue.map((song, index) => (
						<div key={song.id} className={`queue-item ${song.id === nowPlaying?.id ? 'queue-item-active' : ''}`}>
							<div>
								<p className="list-title">{song.title}</p>
								<p className="text-muted">{song.artist || 'Unknown'} · {song.duration || '—'}</p>
							</div>
							<div className="queue-actions">
								<button
									className="btn btn-ghost"
									onClick={() => onMove(song.id, index - 1)}
									disabled={index === 0}
								>
									Up
								</button>
								<button
									className="btn btn-ghost"
									onClick={() => onMove(song.id, index + 1)}
									disabled={index === queue.length - 1}
								>
									Down
								</button>
								<button className="btn btn-ghost" onClick={() => onRemove(song.id)}>
									Remove
								</button>
							</div>
						</div>
					))}
				</div>
			)}
			<div className="divider" />
			<button className="btn btn-secondary" onClick={onSkip}>
				Skip to next
			</button>
		</div>
	);
};

RoomControllerPanel.propTypes = {
	queue: PropTypes.arrayOf(PropTypes.shape({
		id: PropTypes.string.isRequired,
		title: PropTypes.string.isRequired,
		artist: PropTypes.string,
		duration: PropTypes.string,
		thumbnail: PropTypes.string,
	})).isRequired,
	nowPlaying: PropTypes.shape({
		id: PropTypes.string,
		title: PropTypes.string,
		thumbnail: PropTypes.string,
	}),
	onMove: PropTypes.func.isRequired,
	onRemove: PropTypes.func.isRequired,
	onSkip: PropTypes.func.isRequired,
};

export default RoomControllerPanel;
