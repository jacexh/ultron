import React, { useState, useEffect } from 'react';
import { render } from 'react-dom';
import { Line } from '@ant-design/charts';

var COLOR_PLATE_10 = ['#5B8FF9', '#5AD8A6', '#F6BD16', '#E8684A', '#6DC8EC', '#9270CA', '#FF9D4D', '#FF99C3'];

export const LineChart = ({ lineData,localType }) => {
	const [chartOption, setChartOption] = useState({
		data: JSON.parse(localStorage.getItem(localType))||[],
		xField: 'time',
		yField: 'value',
		seriesField: 'category',
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
		point: {
			shape: function shape(_ref) {
				var category = _ref.category;
				return category === 'Gas fuel' ? 'square' : 'circle';
			},
			style: function style(_ref2) {
				var time = _ref2.time;
				return { r: Number(time) % 4 ? 0 : 3 };
			},
		},
	});

	useEffect(() => {
		//只改变data的值比较顺滑流畅
		if (lineData && lineData.length > 0) {
			var newData = chartOption.data.concat(lineData);
			setChartOption({ data: newData });
			localStorage.setItem(localType, JSON.stringify(newData));
		}
	}, [lineData]);

	return <Line {...chartOption} />;
};
