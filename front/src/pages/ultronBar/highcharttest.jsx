import React, { useState, useEffect } from 'react';
import { render } from 'react-dom';
import { Line } from '@ant-design/charts';

const Testdata =[
  {
      "time": "10:18",
      "value": 6,
      "category": "Median Response Time"
  },
  {
      "time": "10:18",
      "value": 5,
      "category": "95% percentile"
  },
  {
    "time": "10:20",
    "value": 16,
    "category": "Median Response Time"
},
{
    "time": "10:20",
    "value": 15,
    "category": "95% percentile"
  },
  {
    "time": "10:21",
    "value": 26,
    "category": "Median Response Time"
},
{
    "time": "10:21",
    "value": 35,
    "category": "95% percentile"
},
]

var COLOR_PLATE_10 = [
  '#5B8FF9',
  '#5AD8A6',
  '#5D7092',
  '#F6BD16',
  '#E8684A',
  '#6DC8EC',
  '#9270CA',
  '#FF9D4D',
  '#269A99',
  '#FF99C3',
];

export const LineChart = (props) => {
  const [data, setData] = useState([]);

  // useEffect(() => {
  //   asyncFetch();
  // }, [])

  // const asyncFetch = () => {
  //   fetch('https://gw.alipayobjects.com/os/bmw-prod/55424a73-7cb8-4f79-b60d-3ab627ac5698.json')
  //     .then((response) => response.json())
  //     .then((json) => setData(json))
  //     .catch((error) => {
  //       console.log('fetch data failed', error);
  //     });
  // };

  var config = {
    data: Testdata,
    xField: 'time',
    yField: 'value',
    seriesField: 'category',
    yAxis: {
      label: {
        formatter: function formatter(v) {
          return ''.concat(v).replace(/\d{1,3}(?=(\d{3})+$)/g, function (s) {
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

  return (
    <Line {...config} />
  )
}



