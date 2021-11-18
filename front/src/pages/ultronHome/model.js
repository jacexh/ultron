import { getMetrics } from '../../../service/chartdataservice';

const Model = {
	namespace: 'home',
	state: {
		metricsStr: '',
	},

	effects: {
		*getMetricsM({ payload, callback }, { call, put, select }) {
			try {
				const response = yield call(getMetrics);
				yield put({
					type: 'setMetrics',
					payload: response,
				});
			} catch (e) {
				console.log(e);
			}
		},
	},

	reducers: {
		setMetrics(state, { payload }) {
			return { ...state, metricsStr: payload };
		},
	},
};

export default Model;
