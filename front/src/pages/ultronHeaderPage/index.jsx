import { useStyles } from '../components/makestyle';
import { ToggleButton, ToggleButtonGroup, Toolbar, Box, AppBar, Divider, IconButton, MoreIcon, StyledFab, Typography, Button } from '@material-ui/core';
import { AccessAlarm, Stop } from '@material-ui/icons';
import { connect } from 'dva';
import styles from './index.css';
import { UltronUsers } from '../ultronUsers/index'
import { useState } from 'react';


export const HeaderStatus = ({ title, textObj, flag = 1, color = '#5E5E5E', openEditUser = 'null' }) => {
  return (
    <>
      <div style={{ paddingLeft: 20, paddingRight: 20 }}>
        <span style={{ fontSize: 14, fontWeight: 600, fontFamily: 'initial', color: '#666666' }}>{title}</span><br />
        <span style={{ fontSize: 17, fontWeight: 600, fontFamily: 'monospace', color: color }}> {textObj}</span><br />
        {title == 'STATUS' ? <a style={{ fontSize: 17, fontWeight: 400, fontFamily: 'monospace', color: '#6495ED' }} onClick={() => openEditUser()}>New Test</a> : ''}
      </div>
      <Divider orientation="vertical" variant="middle" flexItem />
    </>
  )
}

export const UltronHeader = props => {
  const { dispatch } = props;
  const [open, setOpen] = React.useState(true);
  const [flag, setFlag] = useState(true)//是否展示Host
  const [host, setHost] = useState('')


  const handleClose = () => {
    setOpen(false);
  };

  const openEditUser = () => {
    setOpen(true)
    setFlag(false)
  }

  const handleOk = (host, users, spawn) => {
    setHost(host)
    setOpen(false);
  }


  return (
    <>
      <UltronUsers open={open} handleClose={handleClose} handleOk={handleOk} flag={flag} />
      <h1 className={styles.title}>
        <div>
          <AppBar position="fixed" className={useStyles().headerBg}>
            <div>
              <img src="./spaceman.png" width="75" style={{ paddingLeft: 30 }}></img>
              <span style={{ fontSize: 28, fontWeight: 700, paddingLeft: 7, fontFamily: 'fantasy', color: '#404040', }}>Ultron</span>
              <Toolbar className={useStyles().floatRight}>
                <HeaderStatus title='HOST' textObj={host} />
                <HeaderStatus title='STATUS' textObj='RUNNING(100 users)' openEditUser={openEditUser} />
                <HeaderStatus title='RPS' textObj='22.5' />
                <HeaderStatus title='FAILURES' textObj='10%' />
                &nbsp;&nbsp;
                <Button variant="contained" size="large" color="error" startIcon={<Stop />}> STOP</Button>&nbsp;&nbsp;&nbsp;
                <Button variant="contained" size="large">Reset Starts</Button>
              </Toolbar>
            </div>
            <br />
          </AppBar>
        </div >
      </h1 >
    </>
  );
};

