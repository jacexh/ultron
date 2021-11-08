import { useStyles } from '../components/makestyle';
import { Alert, Toolbar, Box, AppBar, Divider, Button } from '@material-ui/core';
import { AccessAlarm, Stop } from '@material-ui/icons';
import { connect } from 'dva';
import styles from './index.css';
import { UltronUsers } from '../ultronUsers/index';
import { useEffect, useState } from 'react';

export const HeaderStatus = ({ title, textObj, flag = 1, color = '#5E5E5E', openEditUser = 'null' }) => {
	return (
		<>
			<div style={{ paddingLeft: 20, paddingRight: 20 }}>
				<span style={{ fontSize: 14, fontWeight: 600, fontFamily: 'initial', color: '#666666' }}>{title}</span>
				<br />
				<span style={{ fontSize: 17, fontWeight: 600, fontFamily: 'monospace', color: color }}> {textObj}</span>
				<br />
				{title == 'STATUS' ? (
					<a style={{ fontSize: 17, fontWeight: 400, fontFamily: 'monospace', color: '#6495ED' }} onClick={() => openEditUser()}>
						New Test
					</a>
				) : (
					''
				)}
			</div>
			<Divider orientation="vertical" variant="middle" flexItem />
		</>
	);
};

export const UltronHeader = ({ getMetrics, metricsStr }) => {
	const [open, setOpen] = useState(true);
	const [metrics, setMetrics] = useState();
	const [stop, setStop] = useState(false);

	useEffect(() => {
		const timerId = setInterval(() => {
			getMetrics();
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
				console.log(res);
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
							<img src="./spaceman.png" width="75" style={{ paddingLeft: 30 }}></img>
							<span style={{ fontSize: 28, fontWeight: 700, paddingLeft: 7, fontFamily: 'fantasy', color: '#404040' }}>Ultron</span>
							<Toolbar className={useStyles().floatRight}>
								{/* <HeaderStatus title='HOST' textObj={host} /> */}
								<HeaderStatus title="STATUS" textObj="RUNNING(100 users)" openEditUser={openEditUser} />
								<HeaderStatus title="TPS" textObj="22.5" />
								<HeaderStatus title="FAILURES" textObj="10%" />
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
