const SEARCH_API_URL = 'https://www.googleapis.com/youtube/v3/search';
const VIDEOS_API_URL = 'https://www.googleapis.com/youtube/v3/videos';

const delay = (ms = 0) => new Promise((resolve) => setTimeout(resolve, ms));

export const searchKaraokeVideos = async (query, apiKey, options = {}) => {
	await delay();
	if (!query) return [];
	const { includeKaraoke = true, orderByViews = false } = options;
	const searchQuery = includeKaraoke ? `${query} karaoke` : query;

	if (!apiKey) {
		return [
			{
				id: `mock-${Date.now()}`,
				videoId: 'dQw4w9WgXcQ',
				title: `${query} (Karaoke Mock)`,
				channelTitle: 'Karayouke Demo',
				thumbnail: 'https://placehold.co/160x160?text=Karaoke',
				duration: '03:30',
			},
		];
	}

	const params = new URLSearchParams({
		part: 'snippet',
		maxResults: '25',
		type: 'video',
		q: searchQuery,
		videoEmbeddable: 'true',
		key: apiKey,
	});

	if (orderByViews) {
		params.set('order', 'viewCount');
	}

	const response = await fetch(`${SEARCH_API_URL}?${params.toString()}`);
	if (!response.ok) {
		throw new Error('Failed to search YouTube videos');
	}
	const data = await response.json();
	const items = data.items || [];
	const ids = items
		.map((item) => item.id?.videoId)
		.filter(Boolean)
		.join(',');

	let durationById = {};
	if (ids) {
		const durationParams = new URLSearchParams({
			part: 'contentDetails',
			id: ids,
			key: apiKey,
		});
		try {
			const durationResponse = await fetch(`${VIDEOS_API_URL}?${durationParams.toString()}`);
			if (durationResponse.ok) {
				const durationData = await durationResponse.json();
				durationById = (durationData.items || []).reduce((acc, item) => {
					acc[item.id] = item.contentDetails?.duration;
					return acc;
				}, {});
			}
		} catch (error) {
			console.warn('Failed to fetch YouTube durations', error);
		}
	}

	return items.map((item) => ({
		id: item.id?.videoId,
		videoId: item.id?.videoId,
		title: item.snippet?.title,
		channelTitle: item.snippet?.channelTitle,
		thumbnail: item.snippet?.thumbnails?.medium?.url,
		duration: durationById[item.id?.videoId],
	}));
};

const formatDuration = (isoDuration) => {
	if (!isoDuration) return null;
	const match = isoDuration.match(/PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?/);
	if (!match) return null;
	const hours = Number(match[1] || 0);
	const minutes = Number(match[2] || 0);
	const seconds = Number(match[3] || 0);
	const totalSeconds = hours * 3600 + minutes * 60 + seconds;
	if (!Number.isFinite(totalSeconds) || totalSeconds <= 0) return null;
	const paddedSeconds = String(totalSeconds % 60).padStart(2, '0');
	const totalMinutes = Math.floor(totalSeconds / 60);
	if (totalMinutes >= 60) {
		const paddedMinutes = String(totalMinutes % 60).padStart(2, '0');
		return `${Math.floor(totalMinutes / 60)}:${paddedMinutes}:${paddedSeconds}`;
	}
	return `${totalMinutes}:${paddedSeconds}`;
};

export const searchYoutube = async (query, options = {}) => {
	const apiKey = import.meta.env.VITE_YOUTUBE_API_KEY;
	const results = await searchKaraokeVideos(query, apiKey, options);
	return results.map((item) => ({
		id: item.videoId || item.id,
		title: item.title,
		artist: item.channelTitle,
		thumbnail: item.thumbnail,
		duration: formatDuration(item.duration) || '—',
	}));
};
