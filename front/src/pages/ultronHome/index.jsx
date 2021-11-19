import React, { useState, useEffect } from 'react';
import { UltronHeader } from '../ultronHeader/index';
import { UltronBar } from '../ultronBar/index';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core';
import { connect } from 'dva';
const moment = require('moment');

const mapStateToProps = state => {
	return {
		home: state.home,
	};
};

const UltronHome = props => {
	const { dispatch } = props;
	const [tableData, setTableData] = useState({});
	const [lineData, setLineData] = useState([]);
	const [tpsLine, setTpsLine] = useState([]);
	const [isPlanEnd, setIsPlanEnd] = useState(false);
	const [metricsTime, setMetricsTime] = useState(moment(new Date()).format('YYYY-MM-DD HH:mm:ss'));
	const { metricsStr } = props.home;

	useEffect(() => {
		getStatistics(metricsStr);
	}, [metricsStr]);

	function getMetricsLength(metricsStr) {
		let maxLength = 0;
		if (metricsStr && metricsStr.length > 0)
			for (var i of metricsStr) {
				if (i.name == 'ultron_attacker_response_time') {
					maxLength = i.metrics.length;
					break;
				}
			}
		return maxLength;
	}

	//获取列表
	function getStatistics(metricsStr) {
		var newStatistic = [];
		var newLineData = [];
		var tpsLineData = [];
		let attacker_length = getMetricsLength(metricsStr);
		attacker_length == 0 ? setIsPlanEnd(true) : '';
		for (var j = 0; j < attacker_length; j++) {
			var optionStatistics = {};
			for (var i of metricsStr) {
				if (i.name == 'ultron_attacker_response_time') {
					if (i['metrics'][j] && i['metrics'][j]['quantiles']) {
						let attacker = i['metrics'][j]['labels']['attacker'];
						var quantiles = i['metrics'][j]['quantiles'];
						optionStatistics.attacker = attacker;
						optionStatistics.MIN = parseFloat(quantiles['0']);
						optionStatistics.P50 = parseFloat(quantiles['0.5']);
						newLineData.push({
							time: metricsTime,
							value: parseFloat(quantiles['0.5']),
							category: attacker + '_50% percentile',
						});
						optionStatistics.P60 = parseFloat(quantiles['0.6']);
						optionStatistics.P70 = parseFloat(quantiles['0.7']);
						optionStatistics.P80 = parseFloat(quantiles['0.8']);
						optionStatistics.P90 = parseFloat(quantiles['0.9']);
						newLineData.push({
							time: metricsTime,
							value: parseFloat(quantiles['0.9']),
							category: attacker + '_90% percentile',
						});
						optionStatistics.P95 = parseFloat(quantiles['0.95']);
						newLineData.push({
							time: metricsTime,
							value: parseFloat(quantiles['0.95']),
							category: attacker + ' 95% percentile',
						});
						optionStatistics.P97 = parseFloat(quantiles['0.97']);
						optionStatistics.P98 = parseFloat(quantiles['0.98']);
						optionStatistics.P99 = parseFloat(quantiles['0.99']);
						newLineData.push({
							time: metricsTime,
							value: parseFloat(quantiles['0.99']),
							category: attacker + '_99% percentile',
						});
						optionStatistics.MAX = parseFloat(quantiles['1']);
					}
				}
				//total request failures tps
				if (i.name == 'ultron_attacker_requests_total') {
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.requests = i['metrics'][j]['value']) : '';
				}
				if (i.name == 'ultron_attacker_failures_total') {
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.failures = i['metrics'][j]['value']) : '';
				}
				//ultron_attacker_tps_total--stop后会显示tps_total，运行中是current tpc --结束标志Plan
				if (i.name == 'ultron_attacker_tps_total') {
					setIsPlanEnd(true);
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.tpsTotal = parseFloat(i['metrics'][j]['value']).toFixed(2)) : '';
				}
				// avg
				if (i.name == 'ultron_attacker_response_time_avg') {
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.AVG = parseFloat(i['metrics'][j]['value'])) : '';
				}
				//current tps
        if (i.name == 'ultron_attacker_tps_current') {
          setIsPlanEnd(false);
					let attacker = i['metrics'][j]['labels']['attacker'];
					tpsLineData.push({
						time: metricsTime,
						value: parseFloat(i['metrics'][j]['value']),
						category: attacker + '_TPS',
					});
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.tpsCurrent = parseFloat(i['metrics'][j]['value']).toFixed(2)) : '';
				}
				//failure ratio 失败率
				if (i.name == 'ultron_attacker_failure_ratio') {
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.failureRatio = Number(i['metrics'][j]['value'] * 100)) : '';
				}
				//users-只有一个
				if (i.name == 'ultron_concurrent_users') {
					i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.users = i['metrics'][0]['value']) : '';
				}
			}
			newStatistic.push(optionStatistics);
    }
		setTableData(newStatistic);
		setLineData(newLineData);
		setTpsLine(tpsLineData);
	}

  function getMetrics() {
		setMetricsTime(moment(new Date()).format('YYYY-MM-DD HH:mm:ss'));
		dispatch({
			type: 'home/getMetricsM',
		});
	}

	return (
		<>
			<UltronHeader getMetrics={getMetrics} metricsStr={metricsStr} tableData={tableData} isPlanEnd={isPlanEnd} />
			<UltronBar tableData={tableData} lineData={lineData} tpsline={tpsLine} />
		</>
	);
};

export default connect(mapStateToProps)(UltronHome);
