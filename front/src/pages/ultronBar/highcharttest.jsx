import React, { useState, useEffect } from 'react';
import { render } from 'react-dom';
import { Line } from '@ant-design/charts';

var COLOR_PLATE_10 = ['#5B8FF9', '#5AD8A6', '#5D7092', '#F6BD16', '#E8684A', '#6DC8EC', '#9270CA', '#FF9D4D', '#269A99', '#FF99C3'];

export const LineChart = ({ lineData }) => {
	const [data, setData] = useState([]);

	useEffect(() => {
    var newdata = data.concat(lineData);
    setData(newdata)
  }, [lineData]);
  
  console.log(data)

	var config = {
		data: data,
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
	};

	return <Line {...config} />;
};
