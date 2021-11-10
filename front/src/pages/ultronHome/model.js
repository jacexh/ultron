import { getMetrics } from '../../../service/chartdataservice';
import parsePrometheusTextFormat from 'parse-prometheus-text-format';
const moment = require('moment');

const Model = {
	namespace: 'home',
	state: {
		metricsStr: '',
		metricsTime:moment(new Date()).format('YYYY-MM-DD HH:mm:ss')
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
			return { ...state, metricsStr: metrics, metricsTime: moment(new Date()).format('HH:mm:ss') };
		},
	},
};

export default Model;
