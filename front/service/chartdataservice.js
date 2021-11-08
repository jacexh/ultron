import request from '../utils/request';

export async function createNewPlan(payload) {
	return request(`/api/v1/plan`, {
		method: 'POST',
		data: payload,
	});
}

export async function getMetrics() {
	return request(`/metrics`, {
		method: 'GET',
	});
}
