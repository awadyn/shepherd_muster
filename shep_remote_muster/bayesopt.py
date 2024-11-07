import math
import random
import subprocess
from subprocess import Popen, PIPE, call
import time
from datetime import datetime
import sys
import getopt
import numpy as np
import itertools
import argparse
import shutil
import os.path

import numpy as np
import itertools
from pprint import pprint
import pandas as pd
import os
from statistics import mean

from ax.plot.contour import plot_contour
from ax.plot.trace import optimization_trace_single_method
from ax.utils.notebook.plotting import render, init_notebook_plotting
from ax.service.managed_loop import optimize
from ax.metrics.branin import branin
from ax.utils.measurement.synthetic_functions import hartmann6

lat_target = 500.0
percentile_target = "99"
TARGET_QPS=100000
MINLATENCY = False
TIME = 10

LATENCIES = {
    "50" : 0.0,
    "75" : 0.0,
    "90" : 0.0,
    "95" : 0.0,
    "99" : 0.0
}

#print(hex2int("0c00"))
    #print(hex2int("1800"))
    #print(int2hexstr(3072))
    #print(int2hexstr(6144))
    
def hex2int(hex_str):
    return int(hex_str,16)

def int2hexstr(int_in):
    s = str(hex(int_in))
    s2 = s[2:]
    if len(s2) == 3:
        s2 = "0"+s2
    assert len(s2) == 4
    return s2
        
def runRemoteCommand(com, server):
    p1 = Popen(["ssh", server, com], stdout=PIPE)
    return p1

def runRemoteCommandGet(com, server):
    p1 = Popen(["ssh", server, com], stdout=PIPE)
    return p1.communicate()[0].strip()

def runRemoteCommands(com, server):
    Popen(["ssh", server, com])
    
def runLocalCommandOut(com):
    #print(com)
    p1 = Popen(list(filter(None, com.strip().split(' '))), stdout=PIPE)
    return p1.communicate()[0].strip()

def runLocalCommands(com):
    Popen(list(filter(None, com.strip().split(' '))))

def runLocalCommand(com):
    p1 = Popen(com, shell=True, stdout=PIPE, stderr=PIPE)
    return p1

def runLocalCommandGet(com):
    p1 = Popen(com, shell=True, stdout=PIPE)
    return p1.communicate()[0].strip()

def img_dnn_eval_func(params):
    global lat_target
    global percentile_target
    global TARGET_QPS

    # launch experiment with params
    itr = str(int(params['itr']))
    dvfs = int2hexstr(int(params['dvfs']))
    print(itr, dvfs)

    #MQPS=200000 MITR=2 MDVFS=0c00 ./run_mcd.sh runStatic
    print(f"MITR={itr} MDVFS={dvfs} MQPS={TARGET_QPS} ./run_mcd.sh runStatic")
    p1 = runLocalCommand(f"MITR={itr} MDVFS={dvfs} MQPS={TARGET_QPS} ./run_mcd.sh runStatic")
    for out in p1.communicate():
        for line in str(out).split("\\n"):
            print(line.strip())
    
    name=f"qps{TARGET_QPS}_itr{itr}_dvfs{dvfs}"
    server_rapl = f"server_rapl_{name}.log"
    mutilate_out = f"mutilate_{name}.out"
    print(name, server_rapl, mutilate_out)

    with open(server_rapl) as file:
        server_rapl_log = [float(line.rstrip()) for line in file]
    #assert len(server_rapl_log) > 25

    joules = 0.0
    lo = 15
    hi = lo+20
    if len(server_rapl_log) > 25:
        print(f"server_rapl_log[{lo}:{hi}]: ", server_rapl_log[lo:hi])
        print("avg_watts ", mean(server_rapl_log[lo:hi]))
        joules = mean(server_rapl_log[lo:hi])
    else:
        # faulty run
        joules = 9999999.0

    with open(mutilate_out) as file:
        for line in file:
            if "read" in line:
                alla = list(filter(None, line.strip().split(' ')))
                LATENCIES["50"] = float(alla[6])
                LATENCIES["90"] = float(alla[7])
                LATENCIES["95"] = float(alla[8])
                LATENCIES["99"] = float(alla[9])

    print(LATENCIES)
    
    old_joules = joules
    ## if ITR, DVFS results in SLA violation then bump up energy use
    if LATENCIES[percentile_target] > lat_target:
        #joules += ((r99th/500.0)*joules)
        #print(joules, joules+(10 * (r99th-505.0)), r99th)
        #joules = joules + (10 * (r99th-505.0))
        ## Joules: 35W -> 140W
        joules = joules * (LATENCIES[percentile_target]-lat_target+1)

    print(f"reward=({percentile_target}% <= {lat_target} us): ITR={itr}, DVFS={dvfs}, Energy={joules} ({old_joules}), Latency={LATENCIES[percentile_target]}")
    print("")
    
    res = {
        'mcd': (joules, 0.0)
    }
    return res

def perform_bayesopt(metric = 'read_99th_mean', minimize = True, ntrials=30):
    global MINLATENCY
    
    results = {}
    counter = 0
    
    itr_vals = [*range(2, 1000, 2)]
    dvfs_vals = [*range(3072, 6656, 1)]
    
    search_space = [{'is_ordered': True,
                     'log_scale': False,
                     'name': 'itr',
                     'type': 'choice',
                     'value_type': 'float',
                     'values': itr_vals},
                    {'is_ordered': True,
                     'log_scale': False,
                     'name': 'dvfs',
                     'type': 'choice',
                     'value_type': 'float',
                     'values': dvfs_vals}]

    #print(f"search space: ITR={len(itr_vals)} DVFS={len(dvfs_vals)}")
    #pprint(search_space)

    if MINLATENCY == True:        
        best_params, values, exp, model = optimize(parameters=search_space,
                                                   evaluation_function=lambda params: img_dnn_eval_func(params),
                                                   experiment_name=f'mcd_discrete',
                                                   objective_name='mcd',
                                                   minimize=minimize,
                                                   total_trials=ntrials)
    else:
        best_params, values, exp, model = optimize(parameters=search_space,
                                                   evaluation_function=lambda params: img_dnn_eval_func(params),
                                                   experiment_name=f'mcd_discrete',
                                                   objective_name='mcd',
                                                   minimize=minimize,
                                                   total_trials=ntrials)
        
    print(f"best_params: itr={str(int(best_params['itr']))} dvfs={int2hexstr(int(best_params['dvfs']))}")
    
    ## set best ITR, DVFS
    #setITR(str(int(best_params['itr'])))
    #setDVFS(str(int(best_params['dvfs'])))
    
if __name__ == '__main__':    
    parser = argparse.ArgumentParser()
    parser.add_argument("--trials", help="Number of trials to run")
    parser.add_argument("--verbose", help="increase output verbosity")
    parser.add_argument("--latency", help="e.g. [200, 250, 350, 450]")
    parser.add_argument("--percentile", help="e.g. [10, 50, 90, 99]")
    parser.add_argument("--qps", required=True, help="Requests-per-second to run bayesopt")
    parser.add_argument("--time", help="time to run")
    parser.add_argument("--minlatency", help="minimize latency")
    
    args = parser.parse_args()
    numTrials = 30
    if args.trials:
        numTrials = int(args.trials)
    if args.latency:
        lat_target = float(int(args.latency))
    if args.percentile:
        percentile_target = args.percentile
    if args.qps:
        TARGET_QPS=int(args.qps)        
    if args.time:
        TIME=int(args.time)
    if args.minlatency:
        MINLATENCY=True
    if args.verbose:
        print(f"numTrials={numTrials}, lat_target={lat_target}, percentile_target={percentile_target}, TARGET_QPS={TARGET_QPS}, TIME={TIME}, MINLATENCY={MINLATENCY}")
        
    perform_bayesopt(ntrials=numTrials)
    
    #com="/root/asplos23/rapl_log.sh "+str(TIME)
    #p2 = Popen(list(filter(None, com.strip().split(' '))), stdout=PIPE)
    #output = p2.communicate()[0].strip().decode('ascii')
    #ll = []
    #print(output)
    #for line in str(output).split("\n"):
    #    ll.append(float(line))
    #    print(line)
    #print(ll[1:len(ll)-1])
    #print(p1.communicate()[0].strip().decode('ascii'))
