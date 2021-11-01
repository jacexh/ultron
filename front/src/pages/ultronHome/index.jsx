import React, { useState, useEffect } from 'react';
import { Tabs, Typography, Tab, Box, TableCell, Paper } from '@material-ui/core';
import { UltronHeader } from '../ultronHeaderPage/index'
import { UltronBar } from '../ultronBar/index'
import { connect } from 'dva';

const mapStateToProps = state => {
  return {
    home: state.home
  }
};


const UltronHome = props => {
  const { form, dispatch } = props;
  const { statisticData } = props.home;
  const [value, setValue] = useState(0);

  const getChartStatic = () => {
    dispatch({
      type: 'home/getChartsStatisticM',
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
