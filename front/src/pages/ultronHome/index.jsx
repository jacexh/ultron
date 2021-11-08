import React, { useState, useEffect } from 'react';
import { UltronHeader } from '../ultronHeader/index';
import { UltronBar } from '../ultronBar/index';
import { connect } from 'dva';

const mapStateToProps = state => {
	return {
		home: state.home,
	};
};

const UltronHome = props => {
	const { dispatch } = props;
	const { metricsStr } = props.home;

	useEffect(() => {
		console.log(metricsStr);
		analysisMetrics(metricsStr);
	}, [metricsStr]);

	function analysisMetrics(metricsStr) {
		for (var i of metricsStr) {
			console.log(i.name);
		}
	}

	function getMetrics() {
		dispatch({
			type: 'home/getMetricsM',
		});
	}

	return (
		<>
			<UltronHeader getMetrics={getMetrics} metricsStr={metricsStr} />
			<UltronBar />
		</>
	);
};

export default connect(mapStateToProps)(UltronHome);
