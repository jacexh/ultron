import React, { useState, useEffect } from 'react';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Tabs, Typography, Tab, Box, Paper } from '@material-ui/core';
import { useStyles } from '../components/makestyle';
import { styled } from '@material-ui/core/styles';
import { LineChart } from '../ultronBar/highcharttest';
import { tableCellClasses } from '@material-ui/core/TableCell';
import { Divider } from 'antd';

const StyledTableCell = styled(TableCell)(({ theme }) => ({
	[`&.${tableCellClasses.head}`]: {
		backgroundColor: theme.palette.common.black,
		color: theme.palette.common.white,
	},
	[`&.${tableCellClasses.body}`]: {
		fontSize: 14,
	},
}));

const StyledTableRow = styled(TableRow)(({ theme }) => ({
	'&:nth-of-type(odd)': {
		backgroundColor: theme.palette.action.hover,
	},
	// hide last border
	'&:last-child td, &:last-child th': {
		border: 0,
	},
}));

function TabPanel(props) {
	const { children, value, index, ...other } = props;
	return (
		<div role="tabpanel" hidden={value !== index} id={`simple-tabpanel-${index}`} aria-labelledby={`simple-tab-${index}`} {...other}>
			{value === index && (
				<Box sx={{ p: 3 }}>
					<Typography>{children}</Typography>
				</Box>
			)}
		</div>
	);
}

function a11yProps(index) {
	return {
		id: `simple-tab-${index}`,
		'aria-controls': `simple-tabpanel-${index}`,
	};
}

export const UltronBar = ({ tableData, lineData, tpsline }, props) => {
	const { dispatch } = props;
	const [value, setValue] = useState(0);

	const handleChange = (event, newValue) => {
		setValue(newValue);
	};

	return (
		<>
			<br />
			<Box sx={{ width: '100%', paddingTop: 4 }}>
				<Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
					<Tabs value={value} onChange={handleChange} centered className={useStyles().root}>
						<Tab label="Statistics" {...a11yProps(0)}></Tab>
						{lineData && lineData.length > 0 ? <Tab label="Charts" {...a11yProps(1)} /> : ''}
					</Tabs>
				</Box>
				<TabPanel value={value} index={0}>
					<TableContainer component={Paper}>
						<Table sx={{ minWidth: 650 }} aria-label="simple table">
							<TableHead>
								<TableRow>
									<StyledTableCell>ATTACKER</StyledTableCell>
									<StyledTableCell align="center">MIN(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P50(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P60(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P70(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P80(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P90(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P95(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P97(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P98(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">P99(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">MAX(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">AVG(ms)&nbsp;</StyledTableCell>
									<StyledTableCell align="center">REQUESTS&nbsp;</StyledTableCell>
									<StyledTableCell align="center">FAILURES&nbsp;</StyledTableCell>
									<StyledTableCell align="center">TPS&nbsp;</StyledTableCell>
								</TableRow>
							</TableHead>
							<TableBody>
								<StyledTableRow>
									<StyledTableCell component="th" scope="row">
										{tableData.attacker}
									</StyledTableCell>
									<StyledTableCell align="center">{tableData.MIN}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P50}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P60}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P70}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P80}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P90}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P95}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P97}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P98}</StyledTableCell>
									<StyledTableCell align="center">{tableData.P99}</StyledTableCell>
									<StyledTableCell align="center">{tableData.MAX}</StyledTableCell>
									<StyledTableCell align="center">{tableData.AVG}</StyledTableCell>
									<StyledTableCell align="center">{tableData.requests}</StyledTableCell>
									<StyledTableCell align="center">{tableData.failures}</StyledTableCell>
									<StyledTableCell align="center">{tableData.tpsCurrent}</StyledTableCell>
								</StyledTableRow>
								<StyledTableRow>
									<StyledTableCell align="center" colSpan={12}></StyledTableCell>
									<StyledTableCell align="center">{tableData ? <span style={{ fontSize: 16, fontWeight: 500 }}>TOTAL</span> : ''}</StyledTableCell>
									<StyledTableCell align="center">{tableData && tableData.requests ? tableData.requests : ''}</StyledTableCell>
									<StyledTableCell align="center">{tableData && tableData.failures ? tableData.failures : ''}</StyledTableCell>
									<StyledTableCell align="center">{tableData && tableData.tpsCurrent ? tableData.tpsCurrent : ''}</StyledTableCell>
								</StyledTableRow>
							</TableBody>
						</Table>
					</TableContainer>
					{/* </Spin> */}
				</TabPanel>
				<TabPanel value={value} index={1}>
					<h3>Response Times(ms)</h3>
					<LineChart lineData={lineData} localType="chartData" />
					<br />
					<br />
					<h3>Total Requests per Second</h3>
					<LineChart lineData={tpsline} localType="tpsline" />
				</TabPanel>
				<TabPanel value={value} index={2}></TabPanel>
				<TabPanel value={value} index={3}>
					Item four
				</TabPanel>
				<TabPanel value={value} index={4}>
					Item five
				</TabPanel>
			</Box>
		</>
	);
};
