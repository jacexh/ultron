import { useStyles } from '../components/makestyle';
import { Alert, Toolbar, Box, AppBar, Divider, Button } from '@material-ui/core';
import { AccessAlarm, Stop } from '@material-ui/icons';
import { connect } from 'dva';
import styles from './index.css';
import { UltronUsers } from '../ultronUsers/index';
import { useEffect, useState } from 'react';

export const HeaderStatus = ({ title, textObj, flag = 1, color = '#5E5E5E', openEditUser = 'null', fontSize = '20' }) => {
	return (
		<>
			<div style={{ paddingLeft: 20, paddingRight: 20 }}>
				<span style={{ fontSize: 14, fontWeight: 600, fontFamily: 'Arial, Helvetica, sans-serif', color: '#666666' }}>{title}</span>
				<br />
				<span style={{ fontSize: 22, fontWeight: 600, fontFamily: 'Arial, Helvetica, sans-serif', color: color }}> {textObj}</span>
				{title == 'PLAN' ? (
					<a style={{ fontSize: 17, fontWeight: 400, fontFamily: 'Arial, Helvetica, sans-serif', color: '#6495ED' }} onClick={() => openEditUser()}>
						Edit
					</a>
				) : (
					''
				)}
			</div>
			<Divider orientation="vertical" variant="middle" flexItem />
		</>
	);
};

export const UltronHeader = ({ getMetrics, tableData }) => {
	const [open, setOpen] = useState(true);
	const [stop, setStop] = useState(false);

	useEffect(() => {
		const timerId = setInterval(() => {
			tableData && tableData.tpsTotal ? '' : getMetrics();
		}, 5000);
		return () => {
			// 组件销毁时，清除定时器
			clearInterval(timerId);
		};
	});

	useEffect(() => {
		getMetrics();
	}, []);

	const handleClose = () => {
		setOpen(false);
	};
	const openEditUser = () => {
		setOpen(true);
	};

	function stopPlan() {
		fetch(`/api/v1/plan`, {
			method: 'DELETE',
		})
			.then(response => response.json())
			.then(function(res) {
				if (res && res.result) {
					setStop(true);
					<Alert severity="success">success</Alert>;
				}
			});
	}

	return (
		<>
			<UltronUsers open={open} handleClose={handleClose} setOpen={setOpen} setStop={setStop} />
			<h1 className={styles.title}>
				<div>
					<AppBar position="fixed" className={useStyles().headerBg}>
						<div>
							<img src="./spaceman.png" width="75" style={{ paddingLeft: 25 }}></img>
							<span style={{ fontSize: 20, fontWeight: 700, paddingLeft: 7, fontFamily: 'fantasy', color: '#404040' }}>Ultron</span>
							<Toolbar className={useStyles().floatRight}>
								<HeaderStatus title="PLAN" openEditUser={openEditUser} />
								<HeaderStatus title="USERS" textObj={tableData && tableData.users ? tableData.users : 0} />
								<HeaderStatus
									title="FAILURES"
									textObj={tableData && tableData.failureRatio ? (parseFloat(tableData.failureRatio) * 100).toFixed(2) + '%' : 0}
								/>
								{tableData && tableData.tpsTotal ? <HeaderStatus title="Total TPS" textObj={tableData.tpsTotal} /> : ''}
								&nbsp;&nbsp;
								{stop ? (
									''
								) : (
									<Button
										variant="contained"
										size="large"
										color="error"
										startIcon={<Stop />}
										onClick={() => {
											stopPlan();
										}}
									>
										STOP
									</Button>
								)}
								&nbsp;&nbsp;&nbsp;
							</Toolbar>
						</div>
						<br />
					</AppBar>
				</div>
			</h1>
		</>
	);
};
