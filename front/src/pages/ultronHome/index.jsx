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
	const [tableData, setTableData] = useState({});
	const [lineData, setLineData] = useState([]);
	const [tpsLine, setTpsLine] = useState([]);
	const { metricsStr, metricsTime } = props.home;

	useEffect(() => {
		getStatistics(metricsStr);
	}, [metricsStr]);

	//获取列表
	function getStatistics(metricsStr) {
		var optionStatistics = {};
		var newLineData = [];
		var tpsLineData = [];
		for (var i of metricsStr) {
			if (i.name == 'ultron_attacker_response_time') {
				if (i['metrics'] && i['metrics'].length > 0) {
					var quantiles = i['metrics'][0]['quantiles']; //只会一个
					optionStatistics.MIN = parseFloat(quantiles['0']);
					optionStatistics.P50 = parseFloat(quantiles['0.5']);
					newLineData.push({
						time: metricsTime,
						value: parseFloat(quantiles['0.5']),
						category: '50% percentile',
					});
					optionStatistics.P60 = parseFloat(quantiles['0.6']);
					optionStatistics.P70 = parseFloat(quantiles['0.7']);
					optionStatistics.P80 = parseFloat(quantiles['0.8']);
					optionStatistics.P90 = parseFloat(quantiles['0.9']);
					newLineData.push({
						time: metricsTime,
						value: parseFloat(quantiles['0.9']),
						category: '90% percentile',
					});
					optionStatistics.P95 = parseFloat(quantiles['0.95']);
					newLineData.push({
						time: metricsTime,
						value: parseFloat(quantiles['0.95']),
						category: '95% percentile',
					});
					optionStatistics.P97 = parseFloat(quantiles['0.97']);
					optionStatistics.P98 = parseFloat(quantiles['0.98']);
					optionStatistics.P99 = parseFloat(quantiles['0.99']);
					newLineData.push({
						time: metricsTime,
						value: parseFloat(quantiles['0.99']),
						category: '99% percentile',
					});
					optionStatistics.MAX = parseFloat(quantiles['1']);
				}
			}
			//total request failures tps
			if (i.name == 'ultron_attacker_requests_total') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.requests = i['metrics'][0]['value']) : '';
			}
			if (i.name == 'ultron_attacker_failures_total') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.failures = i['metrics'][0]['value']) : '';
			}
			//ultron_attacker_tps_total--stop后会显示tps_total，运行中是current tpc --结束标志Plan
			if (i.name == 'ultron_attacker_tps_total') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.tpsTotal = parseFloat(i['metrics'][0]['value']).toFixed(2)) : '';
			}
			//avg
			if (i.name == 'ultron_attacker_response_time_avg') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.AVG = parseFloat(i['metrics'][0]['value'])) : '';
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.attacker = i['metrics'][0]['labels']['attacker']) : '';
			}
			//current tps
			if (i.name == 'ultron_attacker_tps_current') {
				tpsLineData.push({
					time: metricsTime,
					value: parseFloat(i['metrics'][0]['value']),
					category: 'TPS',
				});
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.tpsCurrent = parseFloat(i['metrics'][0]['value']).toFixed(2)) : '';
			}
			//failure ratio 失败率
			if (i.name == 'ultron_attacker_failure_ratio') {
				tpsLineData.push({
					time: metricsTime,
					value: parseFloat(i['metrics'][0]['value'] * 100),
					category: 'Failure Ratio',
				});
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.failureRatio = Number(i['metrics'][0]['value'] * 100).toFixed(2)) : '';
			}
			//users
			if (i.name == 'ultron_concurrent_users') {
				i['metrics'] && i['metrics'].length > 0 ? (optionStatistics.users = i['metrics'][0]['value']) : '';
			}
		}
		setTableData(optionStatistics);
		setLineData(newLineData);
		setTpsLine(tpsLineData);
	}

	function getMetrics() {
		dispatch({
			type: 'home/getMetricsM',
		});
	}

	return (
		<>
			<UltronHeader getMetrics={getMetrics} metricsStr={metricsStr} tableData={tableData} />
			<UltronBar tableData={tableData} lineData={lineData} tpsline={tpsLine} />
		</>
	);
};

export default connect(mapStateToProps)(UltronHome);
