#!/usr/bin/env python
"""Set up a Vitess environment for Java client integration tests.

Every shard gets a master instance. For extra instances, use the
tablet-config option. Upon successful start up, the port for VtGate is
written to stdout, and the program waits for a one line user input
before shutting down.

Start up steps include:
- start MySQL instances
- configure keyspace
- start VtTablets and ensure SERVING mode
- start VtGate instance

Usage example:
java_vtgate_test_helper.py --shards=-80,80- \
  --tablet-config='{"rdonly":1, "replica":1}' --keyspace=test_keyspace \
  --vtgate-port=11111

starts 1 VtGate on the specified port and 6 vttablets - 1 master,
replica and rdonly each per shard.
"""

import utils
import json
import optparse
import sys

import environment
import tablet


class Tablet(tablet.Tablet):

  def __init__(self, shard, tablet_type):
    super(Tablet, self).__init__()
    self.shard = shard
    self.type = tablet_type


class TestEnv(object):

  def __init__(self, options):
    self.keyspace = options.keyspace
    self.schema = options.schema
    self.vschema = options.vschema
    self.vtgate_port = options.vtgate_port
    self.dbname_override = options.dbname_override
    self.tablets = []
    tablet_config = json.loads(options.tablet_config)
    for shard in options.shards.split(','):
      self.tablets.append(Tablet(shard, 'master'))
      for tablet_type, count in tablet_config.iteritems():
        for _ in range(count):
          self.tablets.append(Tablet(shard, tablet_type))

  def set_up(self):
    try:
      environment.topo_server().setup()
      utils.wait_procs([t.init_mysql() for t in self.tablets])
      utils.run_vtctl(['CreateKeyspace', self.keyspace])
      utils.run_vtctl(
          ['SetKeyspaceShardingInfo', '-force', self.keyspace, 'keyspace_id',
           'uint64'])
      for t in self.tablets:
        t.init_tablet(t.type, keyspace=self.keyspace, shard=t.shard)
      utils.run_vtctl(['RebuildKeyspaceGraph', self.keyspace], auto_log=True)
      for t in self.tablets:
        dbname = 'vt_' + self.keyspace

        if self.dbname_override:
          dbname = self.dbname_override

        t.create_db(dbname)
        t.start_vttablet(
            wait_for_state=None,
            extra_args=['-queryserver-config-schema-reload-time', '1',
                        '-init_db_name_override', dbname],
        )
      for t in self.tablets:
        t.wait_for_vttablet_state('SERVING')
      for t in self.tablets:
        if t.type == 'master':
          utils.run_vtctl(
              ['InitShardMaster', self.keyspace+'/'+t.shard, t.tablet_alias],
              auto_log=True)
      utils.run_vtctl(['RebuildKeyspaceGraph', self.keyspace], auto_log=True)
      if self.schema:
        utils.run_vtctl(['ApplySchema', '-sql', self.schema, self.keyspace])
      if self.vschema:
        if self.vschema[0] == '{':
          utils.run_vtctl(['ApplyVSchema', '-vschema', self.vschema])
        else:
          utils.run_vtctl(['ApplyVSchema', '-vschema_file', self.vschema])
      utils.VtGate(port=self.vtgate_port).start(
          cache_ttl='500s',
      )
    except:
      self.shutdown()
      raise

  def shutdown(self):
    # Explicitly kill vtgate first because
    # StreamingServerShutdownIT.java expects an EOF from the vtgate
    # client and not an error that vttablet killed the query (which is
    # seen when vtgate is killed last).
    if utils.vtgate:
      utils.vtgate.kill()
    tablet.kill_tablets(self.tablets)
    teardown_procs = [t.teardown_mysql() for t in self.tablets]
    utils.wait_procs(teardown_procs, raise_on_error=False)
    environment.topo_server().teardown()
    utils.kill_sub_processes()
    utils.remove_tmp_files()
    for t in self.tablets:
      t.remove_tree()


def parse_args():
  global options, args
  parser = optparse.OptionParser(usage='usage: %prog [options]')
  parser.add_option('--shards', action='store', type='string',
                    help="comma separated list of shard names, e.g: '-80,80-'")
  parser.add_option(
      '--tablet-config', action='store', type='string',
      help='json config for for non-master tablets. e.g '
      "{'replica':2, 'rdonly':1}")
  parser.add_option('--keyspace', action='store', type='string')
  parser.add_option('--dbname-override', action='store', type='string')
  parser.add_option('--schema', action='store', type='string')
  parser.add_option('--vschema', action='store', type='string')
  parser.add_option('--vtgate-port', action='store', type='int')
  utils.add_options(parser)
  (options, args) = parser.parse_args()
  utils.set_options(options)


def main():
  env = TestEnv(options)
  env.set_up()
  sys.stdout.write(json.dumps({
      'port': utils.vtgate.port,
      }) + '\n')
  sys.stdout.flush()
  raw_input()
  env.shutdown()

if __name__ == '__main__':
  parse_args()
  main()
