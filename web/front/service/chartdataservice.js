import request from '../utils/request';

export async function getStatisticList() {
  return request(`/stats/requests`, {
    method: 'GET',
  });
}