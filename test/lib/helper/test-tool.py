#!/usr/bin/env python

import argparse
import os
import sys
import time

def cmd_chmtime(*argv):
	parser = argparse.ArgumentParser(
		prog = 'test-tool chmtime',
		description = 'Read or change modification time of files',
	)
	parser.add_argument('-v', '--verbose', action='store_true', help='verbose mode')
	parser.add_argument('-g', '--get', action='store_true', help='get mode')
	parser.add_argument('files', nargs='+', help='files')
	parsed = parser.parse_args(argv)
	ret = 0

	# The first arg may be a timespec
	timespec = parsed.files[0]
	set_eq = 0
	set_time = 0
	if timespec[0] == '=':
		set_eq = 1
		timespec = timespec[1:]
		if timespec[0] == '+':
			set_eq = 2 # relative "in the future"
			timespec = timespec[1:]
	try:
		set_time = int(timespec)
	except ValueError:
		timespec = None
	else:
		parsed.files.pop(0)
	if (set_eq != 0 and set_time < 0) or set_eq == 2:
		now = time.time()
		set_time += int(now)

	# Must have a valid timespec if not in get or verbose mode
	if (not parsed.get and not parsed.verbose) and timespec is None:
		sys.stderr.write("ERROR: not a base-10 integer: %s\n\n" % parsed.files[0])
		parser.print_help()
		return 1
	if len(parsed.files) == 0:
		parser.print_help()
		return 1

	for file in parsed.files:
		st = None
		ts = 0
		try:
			st = os.stat(file)
		except:
			sys.stderr.write("ERROR: failed to stat %s, skipping\n" % file)
			ret = 1
			continue
		if set_eq == 0:
			ts = set_time + int(st.st_mtime)
		else:
			ts = set_time
		if parsed.get:
			print("%ld" % ts)
		elif parsed.verbose:
			print("{ts}\t{file}".format(ts=ts, file=file))
		if timespec is not None and ts != int(st.st_mtime):
			try:
				os.utime(file, (st.st_atime, ts))
			except:
				sys.stderr.write("ERROR: failed to modify time on %s\n" % file)
				ret = 1
	return ret

def cmd_date(*argv):
	parser = argparse.ArgumentParser(
		prog = 'test-tool date',
		description = 'Datetime related utils',
	)
	parser.add_argument('cmd', help='command name')
	parser.add_argument('args', nargs='*', help='arguments')
	parsed = parser.parse_args(argv)
	ret = 0

	if parsed.cmd == 'getnanos':
		seconds = time.time_ns() / 1e9
		if len(parsed.args) > 0:
			seconds -= float(parsed.args[0])
		print("%lf" % seconds)
	elif parsed.cmd == 'is64bit':
		return 1
	elif parsed.cmd == 'time_t-is64bit':
		return 1
	elif parsed.cmd in ('relative', 'human', 'parse', 'approxidate', 'timestamp'):
		raise Exception("Not implemented")
	elif parsed.cmd.startswith('show:'):
		raise Exception("Not implemented")
	else:
		sys.stderr.write("ERROR: unknown date cmd: %s\n" % parsed.cmd)
		return 1
	return 0

def cmd_env_helper(*argv):
	parser = argparse.ArgumentParser(
		prog = 'test-tool env-helper',
		description = 'Read or change modification time of files',
	)
	parser.add_argument('--type', choices=['bool', 'ulong'], required=True, help='value is given this type')
	parser.add_argument('--default', help='default for git_env_*(...) to fall back on')
	parser.add_argument('--exit-code', action='store_true', help='be quiet only use git_env_*() value as exit code')
	parser.add_argument('key', nargs=1, help='environment key')
	parsed = parser.parse_args(argv)
	ret = 0

	def git_parse_maybe_bool_text(val):
		if val is None:
			return 1
		elif val == "":
			return 0
		elif isinstance(val, str) and val.lower() in ('1', 'true', 'yes', 'on'):
			return 1
		elif isinstance(val, str) and val.lower() in ('0', 'false', 'no', 'off'):
			return 0
		else:
			return -1

	def git_parse_maybe_bool(val):
		v = git_parse_maybe_bool_text(val)
		if v >= 0:
			return v
		try:
			v = int(val)
		except ValueError:
			sys.stderr.write("ERROR: bad boolean value '%s'\n" % val)
			exit(128)
		if v == 0:
			return 0
		else:
			return 1

	if parsed.default is not None and len(parsed.default) == 0:
		parser.print_help()
		return 129
	if parsed.type is None:
		parser.print_help()
		return 129

	if parsed.type == 'bool':
		default_int = 0
		if parsed.default is not None:
			default_int = git_parse_maybe_bool(parsed.default)
			if default_int < 0:
				sys.stderr.write("ERROR: option `--default' expects a boolean value with `--type=bool`, not `%s`\n" %
						 parsed.default)
				return 129
		ret = git_parse_maybe_bool(os.getenv(parsed.key[0], default=str(default_int)))
		if not parsed.exit_code:
			if ret == 0:
				print('false')
			else:
				print('true')
	elif parsed.type == 'ulong':
		val = os.getenv(parsed.key[0], default=parsed.default)
		ret = 0
		if val is None or val == "":
			val = "0"
		# Support unit factor like: k/m/g
		val = val.lower()
		try:
			if val[-1] in ('k','m', 'g'):
				unit = val[-1]
				val = val[:-1]
				ret = int(val) * 1024 ** {'k': 1,'m': 2, 'g': 3}[unit]
			else:
				ret = int(val)
		except ValueError:
			sys.stderr.write("ERROR: failed to parse ulong number '%s'\n" % val)
			return 129
		if not parsed.exit_code:
			print('%ld' % ret)

	if ret == 0:
		return 1
	return 0

def cmd_hexdump(*argv):
	have_data = False
	size = 1024
	while True:
		buf = sys.stdin.buffer.read(size)
		if len(buf) == 0:
			break
		have_data = True
		for c in buf:
			print("%02x" % c, end=' ')
	if have_data:
		print()

def cmd_path_utils(*argv):
	parser = argparse.ArgumentParser(
		prog = 'test-tool path-utils',
		description = 'Path related utils',
	)
	parser.add_argument('cmd', help='command name')
	parser.add_argument('args', nargs='*', help='arguments')
	parsed = parser.parse_args(argv)
	ret = 0

	if parsed.cmd == 'file-size':
		for file in parsed.args:
			try:
				st = os.stat(file)
			except:
				sys.stderr.write("ERROR: cannot stat '%s'\n" % file)
				ret = 1
			else:
				print("%ld" % st.st_size)
	elif parsed.cmd == 'skip-n-bytes':
		if len(parsed.args) < 2:
			sys.stderr.write("ERROR: need file and number of bytes to skip\n")
			return 1
		file = parsed.args[0]
		try:
			offset = int(parsed.args[1])
		except ValueError:
			sys.stderr.write("ERROR: '%s' is not a number, wrong order of args?\n" % parsed.args[1])
			return 1
		stream = None
		try:
			stream = open(file, 'rb')
		except:
			sys.stderr.write("ERROR: could not open '%s'\n" % file)
			return 1
		stream.seek(offset,  os.SEEK_SET)
		size=1024
		while True:
			buf = stream.read(size)
			if len(buf) == 0:
				break
			sys.stdout.buffer.write(buf)
	elif parsed.cmd in ('normalize_path_copy',
			    'real_path',
			    'absolute_path',
			    'longest_ancestor_length',
			    'prefix_path',
			    'strip_path_suffix',
			    'print_path',
			    'relative_path',
			    'basename',
			    'dirname',
			    'is_dotgitmodules',
			    'is_dotgitignore',
			    'is_dotgitattributes',
			    'is_dotmailmap',
			    'slice-tests',
			    'protect_ntfs_hfs',
			    'is_valid_path',
			    ):
		raise Exception("Not implemented")
	else:
		sys.stderr.write("ERROR: unknown path-utils cmd: %s\n" % parsed.cmd)
		return 1

	return ret

class Packet:
	MaxPacket = 65520

	PACKET_READ_EOF = 0
	PACKET_READ_NORMAL = 1
	PACKET_READ_FLUSH = 2
	PACKET_READ_DELIM = 3
	PACKET_READ_RESPONSE_END = 4

	hextable = {
		ord('0'): 0,
		ord('1'): 1,
		ord('2'): 2,
		ord('3'): 3,
		ord('4'): 4,
		ord('5'): 5,
		ord('6'): 6,
		ord('7'): 7,
		ord('8'): 8,
		ord('9'): 9,
		ord('a'): 10,
		ord('b'): 11,
		ord('c'): 12,
		ord('d'): 13,
		ord('e'): 14,
		ord('f'): 15,
		ord('A'): 10,
		ord('B'): 11,
		ord('C'): 12,
		ord('D'): 13,
		ord('E'): 14,
		ord('F'): 15,
	}

	def packet_flush(self):
		sys.stdout.buffer.write(b"0000")

	def packet_delim(self):
		sys.stdout.buffer.write(b"0001")

	def packet_response_end(self):
		sys.stdout.buffer.write(b"0002")

	def packet_write(self, data, prefix=None):
		stream = sys.stdout.buffer
		size = len(data) + 4
		if prefix is not None:
			size+=1
		stream.write(b"%04x" % size)
		if prefix is not None:
			stream.write(b"%c" % prefix)
		stream.write(data)

	def pack_line(self, line):
		if line == b"0000" or line == b"0000\n":
			self.packet_flush()
		elif line == b"0001" or line == b"0001\n":
			self.packet_delim()
		elif line == b"0002" or line == b"0002\n":
			self.packet_response_end()
		else:
			self.packet_write(line)

	def packet_reader_read(self):
		stream = sys.stdin.buffer
		head = stream.read(4)
		if len(head) == 0:
			return None, self.PACKET_READ_EOF
		elif len(head) != 4:
			raise Exception('Short header: %s' % head)
		if head == b"0000":
			return None, self.PACKET_READ_FLUSH
		elif head == b"0001":
			return None, self.PACKET_READ_DELIM
		elif head == b"0002":
			return None, self.PACKET_READ_RESPONSE_END
		elif head in (b"0003", b"0004"):
			raise Exception('Invalid header: %s' % str(head, 'utf-8'))
		size = 0
		for c in head:
			size = (size << 4) | self.hextable[c]
		size -= 4
		buffer = stream.read(size)
		if len(buffer) != size:
			raise Exception('short read (%d != %d)' % (len(buffer), size))
		return buffer, self.PACKET_READ_NORMAL

	def unpack(self):
		while True:
			buffer, status = self.packet_reader_read()
			if status == self.PACKET_READ_EOF:
				break
			elif status == self.PACKET_READ_FLUSH:
				sys.stdout.buffer.write(b"0000\n")
			elif status == self.PACKET_READ_DELIM:
				sys.stdout.buffer.write(b"0001\n")
			elif status == self.PACKET_READ_RESPONSE_END:
				sys.stdout.buffer.write(b"0002\n")
			elif status == self.PACKET_READ_NORMAL:
				sys.stdout.buffer.write(buffer)
			else:
				raise Exception("Unknown status")

def cmd_pkt_line(*argv):
	parser = argparse.ArgumentParser(
		prog = 'test-tool pkt-line',
		description = 'Packet related utils',
	)
	parser.add_argument('cmd', help='command name')
	parser.add_argument('args', nargs='*', help='arguments')
	parsed = parser.parse_args(argv)
	packet = Packet()
	ret = 0

	if parsed.cmd == 'pack':
		if len(parsed.args) == 0:
			while True:
				line = sys.stdin.buffer.readline()
				if len(line) == 0:
					break
				packet.pack_line(line)
		else:
			for line in parsed.args:
				packet.pack_line(bytes(line, 'UTF-8'))
	elif parsed.cmd == 'pack-raw-stdin':
		while True:
			data = sys.stdin.buffer.read(Packet.MaxPacket - 4)
			if len(data) == 0:
				break
			packet.packet_write(data)
	elif parsed.cmd == 'unpack':
		packet.unpack()
	elif parsed.cmd in ('unpack-sideband',
			    'send-split-sideband',
			    'receive-sideband',
			    ):
		raise Exception("Not implemented")
	else:
		sys.stderr.write("ERROR: unknown pkt-line cmd: %s\n" % parsed.cmd)
		return 1

def cmd_xml_encode(*argv):
	size = 1024
	while True:
		buf = sys.stdin.read(size)
		if len(buf) == 0:
			break
		for ch in buf:
			c = ord(ch)
			if ch == '&':
				sys.stdout.write('&amp;')
			elif ch == '\'':
				sys.stdout.write('&apos;')
			elif ch == '"':
				sys.stdout.write('&quot;')
			elif ch == '<':
				sys.stdout.write('&lt;')
			elif ch == '>':
				sys.stdout.write('&gt;')
			elif c >= 0x20:
				sys.stdout.write(ch)
			elif c == 0x09 or c == 0x0a or c == 0x0d:
				sys.stdout.write('&#x%02x;' % c)

commands = [
	    {'name': 'chmtime',		'fn': cmd_chmtime},
	    {'name': 'date',		'fn': cmd_date},
	    {'name': 'env-helper',	'fn': cmd_env_helper},
	    {'name': 'hexdump',		'fn': cmd_hexdump},
	    {'name': 'path-utils',	'fn': cmd_path_utils},
	    {'name': 'pkt-line',	'fn': cmd_pkt_line},
	    {'name': 'xml-encode',	'fn': cmd_xml_encode},
	   ]

def usage():
	sys.stderr.write('''
Usage: test-tool <cmd> ...

NOTE: You can also extend test-tool cmd by writing your own extensions
      inside module "test-tools" in your test directory. E.g.:
      Write your sub-command in "test-tools/sub-cmd.py".
''')
	exit(1)

if len(sys.argv) < 2:
	usage()

cmd_name, args = sys.argv[1], sys.argv[2:]

for cmd in commands:
	if cmd_name == cmd['name']:
		ret = cmd['fn'](*args)
		exit(ret)

# When 'test-tool' is called from test framework, the TEST_DIRECTORY env will
# always be set. Users could develop their own test-tool extensions as module
# inside <TEST_DIRECTORY>, such as '<TEST_DIRECTORY>/test-tools/foo-bar.py'.
# By this way, users could run command "test-tool foo-bar ..." in test cases.
#
# Find specific method (such as "Run()") in a module. User can write module
# with arbitrary filename, such as "foo-bar" instead of "foo_bar", so import
# module using "__import__" and getattr().

def get_method_from_module(module, names):
	if module is None:
		module = __import__('.'.join(names[:-1]))
		names.pop(0)
	elif len(names) > 1:
		module = getattr(module, names.pop(0))
	elif len(names) == 1:
		return getattr(module, names.pop(0))
	return get_method_from_module(module, names)

# Try to find extension module in <TEST_DIRECTORY>, so we need add it
# to package search list defined by "sys.path".
test_dir = os.getenv("TEST_DIRECTORY")
if test_dir is None or test_dir == '':
	sys.stderr.write('WARN: env TEST_DIRECTORY is not set, '
			 'unknown command: %s\n' % cmd_name)
	exit(1)

# Check existence of <TEST_DIRECTORY>/test-tools.
if not os.path.isdir(os.path.join(test_dir, "test-tools")):
	sys.stderr.write('ERROR: unknown command: %s\n' % cmd_name)
	exit(1)

# Check whether <TEST_DIRECTORY>/test-tools is a valid module.
if not os.path.isfile(os.path.join(test_dir, "test-tools", "__init__.py")):
	sys.stderr.write('ERROR: test-tools is not a valid module, '
			 'must have a "__init__.py" inside a module\n')
	exit(1)

# Check existence <TEST_DIRECTORY>/test-tools/<cmd>.py.
if not os.path.isfile(os.path.join(test_dir, "test-tools", "%s.py" % cmd_name)):
	sys.stderr.write('ERROR: cannot find "test-tools/%s.py", '
			 'unknown command: %s\n' % (cmd_name, cmd_name))
	exit(1)

# Load extension module and execute "Run()" method.
sys.path.append(test_dir)
try:
	run = get_method_from_module(None, ['test-tools', cmd_name, 'Run'])
except ModuleNotFoundError:
	sys.stderr.write('ERROR: fail to find module for command: %s\n' % cmd_name)
except AttributeError:
	sys.stderr.write('ERROR: must have a "Run()" method in module '
                     'to support command: %s\n' % cmd_name)
except:
    raise
else:
    # No exception, run method
	run(*args)
