#!/usr/bin/env python
# -*- coding: utf-8 -*-
# author： angus time:2024/5/27
from datetime import datetime

import pendulum
import json
from airflow.models import DAG
from airflow.operators.bash import BashOperator

log_dir = '/data/logs/x-project1'
local_tz = pendulum.timezone("Asia/Shanghai")
start_date = datetime(2024, 1, 1, tzinfo=local_tz)
bash_command = """ cd /data/services/x-project1 && ./service {runner_file} """

default_args = {
    'owner': 'airflow',
    'email': [""],
    'email_on_failure': True,
    'max_active_runs': 1,
    'params': {
        'log_dir': log_dir,
    }
}

dag_submit_order = DAG(
    dag_id='dag_submit_order',
    default_args=default_args,
    start_date=start_date,
    schedule_interval='30 11 * * MON-FRI',
    catchup=False,
    tags=['x-project1', 'tob submit bond order task']
)

dag_submit_order_task = BashOperator(
    task_id='dag_submit_order_task',
    dag=dag_submit_order,
    bash_command=bash_command.format(runner_file="--job=submit_order_task")
)