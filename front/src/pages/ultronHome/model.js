import {
  getStatisticList,
} from '../../../service/chartdataservice';
import { message } from 'antd';

const Model = {
  namespace: 'home',
  state: {
    statisticData: '212'
  },

  effects: {
    *getChartsStatisticM({ }, { call, put, select }) {
      console.log('aa')
      try {
        // 对接ljdp后端登录
        const response = yield call(getStatisticList);
        if (response.code == 200) {
          yield put({
            type: 'setStatisticList',
            payload: response,
          });
        }
      } catch (e) {
        console.log(e)
      }
    }
  },

  reducers: {
    setStatisticList(state, { payload }) {
      return { ...state, statisticData: payload.data.list };
    },
  },
}

export default Model;