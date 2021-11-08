import { getMetrics } from '../../../service/chartdataservice';
import parsePrometheusTextFormat from 'parse-prometheus-text-format';

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
      const metrics = parsePrometheusTextFormat(payload);
			return { ...state, metricsStr: metrics };
		},
	},
};

export default Model;
