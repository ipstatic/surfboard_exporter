import re
import time
from prometheus_client.core import GaugeMetricFamily, CounterMetricFamily
from lxml import html,etree

class SurfboardCollector():
    def collect(self):
        """
        Screenscape a Surfboard modem and return Prometheus text format metrics
        """
        start = time.time()

        page = html.parse('http://192.168.100.1/cgi-bin/status')
        downstream = page.xpath('//table')[2]
        upstream = page.xpath('//table')[3]

        ds_frequency = GaugeMetricFamily('surfboard_downstream_frequency_megahertz', 'Downstream frequency in Megahertz', labels=['channel'])
        ds_power = GaugeMetricFamily('surfboard_downstream_power_dbmv', 'Downstream power level in dBmv', labels=['channel'])
        ds_snr = GaugeMetricFamily('surfboard_downstream_snr_db', 'Downstream signal to noise ration in dB', labels=['channel'])
        ds_codewords = GaugeMetricFamily('surfboard_downstream_codewords_total', 'Downstream codewords (corrected, uncorrectable)', labels=['channel', 'corrrected', 'uncorrectable'])
        us_frequency = GaugeMetricFamily('surfboard_upstream_frequency_megahertz', 'Upstream frequency in Megahertz', labels=['channel'])
        us_power = GaugeMetricFamily('surfboard_upstream_power_dbmv', 'Upstream power level in dBmv', labels=['channel'])

        for i, row in enumerate(downstream):
            # Have to skip header row, wish this was in thead instead of tbody
            if i > 1:
                channel = row.xpath('td[4]')[0].text

                value = float(re.findall('(\d+\.\d+)', row.xpath('td[5]')[0].text)[0])
                ds_frequency.add_metric([channel], value)
                
                value = float(re.findall('(\d+\.\d+)', row.xpath('td[6]')[0].text)[0])
                ds_power.add_metric([channel], value)

                value = float(re.findall('(\d+\.\d+)', row.xpath('td[7]')[0].text)[0])
                ds_snr.add_metric([channel], value)

                corrected = int(row.xpath('td[8]')[0].text)
                uncorrectable = int(row.xpath('td[9]')[0].text)
                value = int(corrected + uncorrectable)
                ds_codewords.add_metric([channel, str(corrected), str(uncorrectable)], value)

        for i, row in enumerate(upstream):
            # Have to skip header row, wish this was in thead instead of tbody
            if i > 1:
                channel = row.xpath('td[4]')[0].text

                value = float(re.findall('(\d+\.\d+)', row.xpath('td[6]')[0].text)[0])
                us_frequency.add_metric([channel], value)
                
                value = float(re.findall('(\d+\.\d+)', row.xpath('td[7]')[0].text)[0])
                us_power.add_metric([channel], value)

        yield ds_frequency
        yield ds_power
        yield ds_snr
        yield ds_codewords
        yield us_frequency
        yield us_power
        yield GaugeMetricFamily('surfboard_scrape_duration_seconds', 'Time Surfboard scrape took, in seconds', value=(time.time() - start))
