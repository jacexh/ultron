import React, { useState, useEffect } from 'react';
import { render } from 'react-dom';
import HighchartsReact from 'highcharts-react-official';
import Highcharts from 'highcharts';

export const LineChart = (props) => {
  const [option, setOption] = useState({
    chartOptions: {
      xAxis: {
        categories: ['A', 'B', 'C'],
      },
      series: [{ data: [1, 2, 3] }],
      // plotOptions: {
      //   series: {
      //     point: {
      //       events: {
      //         mouseOver: setHoverData
      //       }
      //     }
      //   }
      // }
    },
    hoverData: null,
  });

  const setHoverData = (e) => {
    // The chart is not updated because `chartOptions` has not changed.
    setOption({ hoverData: e.target.category });
  };

  const updateSeries = () => {
    // The chart is updated only with new options.
    setOption({
      chartOptions: {
        series: [{ data: [Math.random() * 5, 2, 1] }],
      },
    });
  };



  return (
    <div>
      <HighchartsReact highcharts={Highcharts} options={option.chartOptions} />
      <h3>Hovering over {option.hoverData}</h3>
      <button onClick={() => updateSeries()}>Update Series</button></div>
  )

}



