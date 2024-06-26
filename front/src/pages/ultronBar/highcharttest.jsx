import React, { useState, useEffect } from 'react';
import { render } from 'react-dom';
import { Line } from '@ant-design/charts';
const moment = require('moment');

var COLOR_PLATE_10 = ['#5B8FF9', '#5AD8A6', '#F6BD16', '#E8684A', '#6DC8EC', '#9270CA', '#FF9D4D', '#FF99C3'];

export const LineChart = ({ lineData, localType }) => {
	const [chartOption, setChartOption] = useState({
		data: [],
		xField: 'time',
		yField: 'value',
		seriesField: 'category',
		xAxis: {
			label: {
				formatter: function(v) {
					return moment(v).format('HH:mm:ss');
				},
			},
		},
		yAxis: {
			label: {
				formatter: function formatter(v) {
					return ''.concat(v).replace(/\d{1,3}(?=(\d{3})+$)/g, function(s) {
						return ''.concat(s, ',');
					});
				},
			},
		},
		color: COLOR_PLATE_10,
		legend: {
			position: 'top',
		},
		smooth: true,
		// @TODO 后续会换一种动画方式
		animation: {
			appear: {
				animation: 'path-in',
				duration: 5000,
			},
		},
	});

	useEffect(() => {
		//只改变data的值比较顺滑流畅
		if (localStorage.getItem(localType)) {
			var newData = [];
			JSON.parse(localStorage.getItem(localType)).map(function(item) {
				let d1 = new Date(item.time.replace(/\-/g, '/'));
				let d2 = new Date();
				let newTime = getDiffTime(d2, d1);
				if (newTime < 15) newData.push(item); //保留15分钟之内的
			});
			let chartData = newData.concat(lineData);
			localStorage.setItem(localType, JSON.stringify(chartData));
			setChartOption({ data: chartData });
		} else {
			//初始化
			localStorage.setItem(localType, JSON.stringify(lineData));
			setChartOption({ data: lineData });
		}
	}, [lineData]);

	function getDiffTime(d2, d1) {
		let disparity = d2.getTime() - d1.getTime();
		let min = Math.round(disparity / 1000 / 60);
		return min;
	}

	return <Line {...chartOption} />;
};
