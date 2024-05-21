import request from '../utils/request';

export async function getMetrics() {
	return request(`/metrics.json`, {
		method: 'GET',
	});
}
