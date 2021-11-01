import React, { useState, useEffect } from 'react';
import { Tabs, Typography, Tab, Box, TableCell, Paper } from '@material-ui/core';
import { UltronHeader } from '../ultronHeaderPage/index'
import { UltronBar } from '../ultronBar/index'
import { connect } from 'dva';

const mapStateToProps = state => {
  // const home = state['home'];
  return {
    home: state.home
  }
};


const UltronHome = props => {
  const { form, dispatch } = props;
  console.log(dispatch)
  const { statisticData } = props.home;
  const [value, setValue] = useState(0);

  const getChartStatic = () => {
    // alert(statisticData)
    dispatch({
      type: 'home/getChartsStatisticM',
      // payload: { a: 1 }
    });
  }


  const handleChange = (event, newValue) => {
    setValue(newValue);
  };

  return (
    <>
      <UltronHeader />
      <UltronBar getChartStatic={getChartStatic} />
    </>
  );
};

export default connect(mapStateToProps)(UltronHome)
