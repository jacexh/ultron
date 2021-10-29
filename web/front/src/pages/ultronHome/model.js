import {
  getStatisticList,
} from '../../../service/chartdataservice';
import { message } from 'antd';

const Model = {
  namespace: 'home',
  state: {
    statisticData: '212'
  },

  effetcs: {
    *getChartsStatisticM() {
      alert('aa')
    }
  },

  //   *getChartsStatisticM({ payload }, { call, put, select }) {
  //     alert('aa')
  //     //   try {
  //     //     // 对接ljdp后端登录
  //     //     alert('aa')
  //     //     const response = yield call(getStatisticList, payload);
  //     //     yield put({
  //     //       type: 'setStatisticList',
  //     //       payload: response,
  //     //     });
  //     //   } catch (e) {
  //     //     // ... 交互
  //     //     console.log(e)
  //     //   }
  //     // }
  //     //   alert(payload)
  //     //   try(
  //     //   const response = yield call(getStatisticList);
  //     //   if (response.code == 200) {
  //     //     // yield put({
  //     //     //   type: 'setStatisticList',
  //     //     //   payload: response,
  //     //     // });
  //     //   }
  //     //   // else message.error(response.error_message + response.error_details);
  //     // ) catch(e) {

  //     // }
  // }),

  reducers: {
    setStatisticList(state, { payload }) {
      return { ...state, statisticData: payload.data.list };
    },
  },
};

export default Model;