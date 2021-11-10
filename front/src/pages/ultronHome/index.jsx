import React, { useState, useEffect } from 'react';
import { UltronHeader } from '../ultronHeader/index';
import { UltronBar } from '../ultronBar/index';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core';
import { connect } from 'dva';

const mapStateToProps = state => {
	return {
		home: state.home,
	};
};

const UltronHome = props => {
	const { dispatch } = props;
	const [tableData, setTableData] = useState([]);
	const [lineData, setLineData] = useState([]);
  const { metricsStr, metricsTime } = props.home;

	useEffect(() => {
		getStatistics(metricsStr);
	}, [metricsStr]);

  //获取列表
  function getStatistics(metricsStr) {
    console.log(metricsStr)
		var newResponstTime = [];
    var optionStatistics = {};
    var newLineData=[]
		for (var i of metricsStr) {
			if (i.name == 'ultron_attacker_response_time') {
				if (i['metrics'] && i['metrics'].length > 0) {
					var quantiles = i['metrics'][0]['quantiles'];
					optionStatistics.MIN = parseFloat(quantiles['0']);
          optionStatistics.P50 = parseFloat(quantiles['0.5']);
          newLineData.push({
            "time": metricsTime,
            "value":parseFloat(quantiles['0.5']),
            "category": "50% percentile"
          })
          optionStatistics.P60 = parseFloat(quantiles['0.6']);
          // newLineData.push({
          //   "time": metricsTime,
          //   "value":parseFloat(quantiles['0.6']),
          //   "category": "60% percentile"
          // })
          optionStatistics.P70 = parseFloat(quantiles['0.7']);
          // newLineData.push({
          //   "time": metricsTime,
          //   "value":parseFloat(quantiles['0.7']),
          //   "category": "70% percentile"
          // })
          optionStatistics.P80 = parseFloat(quantiles['0.8']);
          // newLineData.push({
          //   "time": metricsTime,
          //   "value":parseFloat(quantiles['0.8']),
          //   "category": "80% percentile"
          // })
          optionStatistics.P90 = parseFloat(quantiles['0.9']);
          newLineData.push({
            "time": metricsTime,
            "value":parseFloat(quantiles['0.9']),
            "category": "90% percentile"
          })
          optionStatistics.P95 = parseFloat(quantiles['0.95']);
          newLineData.push({
            "time": metricsTime,
            "value":parseFloat(quantiles['0.95']),
            "category": "95% percentile"
          })
          optionStatistics.P97 = parseFloat(quantiles['0.97']);
          // newLineData.push({
          //   "time": metricsTime,
          //   "value":parseFloat(quantiles['0.97']),
          //   "category": "97% percentile"
          // })
          optionStatistics.P98 = parseFloat(quantiles['0.98']);
          // newLineData.push({
          //   "time": metricsTime,
          //   "value":parseFloat(quantiles['0.98']),
          //   "category": "98% percentile"
          // })
          optionStatistics.P99 = parseFloat(quantiles['0.99']);
          newLineData.push({
            "time": metricsTime,
            "value":parseFloat(quantiles['0.99']),
            "category": "99% percentile"
          })
					optionStatistics.MAX = parseFloat(quantiles['1']);
          newResponstTime.push(optionStatistics);
				}
			}
			if (i.name == 'ultron_attacker_requests_total') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.requests = i['metrics'][0]['value']) : '';
			}
			if (i.name == 'ultron_attacker_failures_total') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.failures = i['metrics'][0]['value']) : '';
			}
			if (i.name == 'ultron_attacker_response_time_avg') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.AVG = parseFloat(i['metrics'][0]['value'])) : '';
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.attacker = i['metrics'][0]['labels']['attacker']) : '';
			}
			if (i.name == 'ultron_attacker_tps_current') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.tpsCurrent = parseFloat(i['metrics'][0]['value']).toFixed(2)) : '';
			}
			if (i.name == 'ultron_attacker_tps_total') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.tpsTotal = parseFloat(i['metrics'][0]['value']).toFixed(2)) : '';
			}
			if (i.name == 'ultron_concurrent_users') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.users = i['metrics'][0]['value']) : '';
			}
    }
    setTableData(newResponstTime);
    setLineData(newLineData)
	}



	function getMetrics() {
		dispatch({
			type: 'home/getMetricsM',
		});
	}

	return (
		<>
			<UltronHeader getMetrics={getMetrics} metricsStr={metricsStr} tableData={tableData} />
      <UltronBar tableData={tableData} lineData={lineData}/>
		</>
	);
};

export default connect(mapStateToProps)(UltronHome);
