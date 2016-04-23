#!/usr/bin/python3

import os
import re
import datetime
import gzip
import shutil

mydate = datetime.datetime.now()
month = mydate.strftime("%B")
rootdir = '/var/gemplog/'

for subdir, dirs, files in os.walk(rootdir):
        for file in files:
                if re.search(month, subdir, re.IGNORECASE):
                        continue # current month should be ignored
                log = os.path.join(subdir, file)
                if re.search(".gz", file):
                        continue # already gzipped this

                with open(log, 'rb') as f_in, gzip.open(log + '.gz', 'wb') as f_out:
                        shutil.copyfileobj(f_in, f_out)
                os.remove(log)
