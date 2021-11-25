import React, { useState, useEffect } from 'react';
import { Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Card, CardHeader, CardContent, Box, Paper } from '@material-ui/core';
import { useStyles } from '../components/makestyle';
import { styled } from '@material-ui/core/styles';
import { LineChart } from '../ultronBar/highcharttest';
import { tableCellClasses } from '@material-ui/core/TableCell';

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

export const UltronBar = ({ tableData, lineData, tpsline }, props) => {
	const { dispatch } = props;
	const [totalRequest, setTotalRequest] = useState(0);
	const [totalFail, setTotalFail] = useState(0);
	const [totalCurrent, setTotalCurrent] = useState(0);

	useEffect(() => {
		getTotalRequest(tableData);
	}, [tableData]);

	function getTotalRequest(tableData) {
		let totalRequests = 0;
		let totalFails = 0;
		let totalCurrents = 0;

		tableData && tableData.length > 0
			? tableData.map(i => {
					totalRequests += parseInt(i.requests);
					totalFails += parseFloat(i.failures);
					i.tpsCurrent ? (totalCurrents += parseFloat(i.tpsCurrent)) : '';
			  })
			: '';
		setTotalRequest(totalRequests);
		setTotalFail(totalFails);
		setTotalCurrent(totalCurrents.toFixed(2));
	}

	return (
		<Card sx={{ paddingTop: 5 }}>
			<CardHeader
				title={
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
								{tableData && tableData.length > 0
									? tableData.map((i, index) => {
											return (
												<StyledTableRow key={index}>
													<StyledTableCell component="th" scope="row">
														{i.attacker}
													</StyledTableCell>
													<StyledTableCell align="center">1</StyledTableCell>
													<StyledTableCell align="center">{i.P50}</StyledTableCell>
													<StyledTableCell align="center">{i.P60}</StyledTableCell>
													<StyledTableCell align="center">{i.P70}</StyledTableCell>
													<StyledTableCell align="center">{i.P80}</StyledTableCell>
													<StyledTableCell align="center">{i.P90}</StyledTableCell>
													<StyledTableCell align="center">{i.P95}</StyledTableCell>
													<StyledTableCell align="center">{i.P97}</StyledTableCell>
													<StyledTableCell align="center">{i.P98}</StyledTableCell>
													<StyledTableCell align="center">{i.P99}</StyledTableCell>
													<StyledTableCell align="center">{i.MAX}</StyledTableCell>
													<StyledTableCell align="center">{i.AVG}</StyledTableCell>
													<StyledTableCell align="center">{i.requests}</StyledTableCell>
													<StyledTableCell align="center">{i.failures}</StyledTableCell>
													<StyledTableCell align="center">{i.tpsCurrent ? i.tpsCurrent : 0}</StyledTableCell>
												</StyledTableRow>
											);
									  })
									: ''}
								<StyledTableRow>
									<StyledTableCell align="center" colSpan={12}></StyledTableCell>
									<StyledTableCell align="center">{tableData ? <span style={{ fontSize: 16, fontWeight: 500 }}>TOTAL</span> : ''}</StyledTableCell>
									<StyledTableCell align="center">{parseInt(totalRequest)}</StyledTableCell>
									<StyledTableCell align="center">{parseFloat(totalFail)}</StyledTableCell>
									<StyledTableCell align="center">{parseFloat(totalCurrent)}</StyledTableCell>
								</StyledTableRow>
							</TableBody>
						</Table>
					</TableContainer>
				}
			></CardHeader>
			<CardContent>
				<h2 style={{ fontFamily: 'Arial, Helvetica, sans-serif' }}>Response Times(ms)</h2>
				<LineChart lineData={lineData} localType="chartData" />
				<br />
				<br />
				<h2 style={{ fontFamily: 'Arial, Helvetica, sans-serif' }}>TPS</h2>
				<LineChart lineData={tpsline} localType="tpsline" />
			</CardContent>
		</Card>
	);
};
